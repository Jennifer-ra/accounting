package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Jennifer-ra/accounting/services/go-api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	defaultExpenseCategories = []string{"餐饮", "交通", "购物", "住房", "娱乐", "医疗", "学习", "人情", "其他"}
	defaultIncomeCategories  = []string{"工资", "奖金", "副业", "理财", "报销", "其他"}
	defaultAccounts          = []string{"默认账户", "微信", "支付宝", "银行卡", "现金"}
	monthPattern             = regexp.MustCompile(`^20\d{2}-(0[1-9]|1[0-2])$`)
	datePattern              = regexp.MustCompile(`^20\d{2}-(0[1-9]|1[0-2])-([0-2]\d|3[01])$`)
)

type AccountingService struct{ db *gorm.DB }

type TransactionInput struct {
	ClientID  string `json:"clientId"`
	Type      string `json:"type" binding:"required"`
	AmountFen int64  `json:"amountFen" binding:"required"`
	Category  string `json:"category" binding:"required"`
	Account   string `json:"account" binding:"required"`
	Date      string `json:"date" binding:"required"`
	Note      string `json:"note"`
	Source    string `json:"source"`
}

type NameInput struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type"`
}

type BudgetInput struct {
	AmountFen int64 `json:"amountFen"`
}

type Summary struct {
	Income        int64          `json:"income"`
	Expense       int64          `json:"expense"`
	Balance       int64          `json:"balance"`
	Count         int64          `json:"count"`
	Budget        int64          `json:"budget"`
	BudgetLeft    int64          `json:"budgetLeft"`
	BudgetPercent int            `json:"budgetPercent"`
	ExpenseRanks  []CategoryRank `json:"expenseRanks"`
	IncomeRanks   []CategoryRank `json:"incomeRanks"`
}

type CategoryRank struct {
	Name      string `json:"name"`
	AmountFen int64  `json:"amountFen"`
}

type Bootstrap struct {
	Book             model.Book          `json:"book"`
	Categories       map[string][]string `json:"categories"`
	Accounts         []string            `json:"accounts"`
	MonthlyBudgetFen int64               `json:"monthlyBudgetFen"`
}

func NewAccountingService(db *gorm.DB) *AccountingService { return &AccountingService{db: db} }

func (s *AccountingService) DefaultBook(userID uint) (model.Book, error) {
	return ensureDefaultBook(s.db, userID)
}

func ensureDefaultBook(db *gorm.DB, userID uint) (model.Book, error) {
	var book model.Book
	err := db.Where("owner_user_id = ?", userID).First(&book).Error
	if err == nil {
		return book, ensureDefaults(db, book.ID, userID)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Book{}, err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		book = model.Book{OwnerUserID: userID, Name: "个人账本"}
		if err := tx.Create(&book).Error; err != nil {
			return err
		}
		member := model.BookMember{BookID: book.ID, UserID: userID, Role: "owner"}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&member).Error; err != nil {
			return err
		}
		return ensureDefaults(tx, book.ID, userID)
	})
	return book, err
}

func ensureDefaults(db *gorm.DB, bookID, userID uint) error {
	for i, name := range defaultExpenseCategories {
		if err := createCategoryIfMissing(db, bookID, "expense", name, i, true); err != nil {
			return err
		}
	}
	for i, name := range defaultIncomeCategories {
		if err := createCategoryIfMissing(db, bookID, "income", name, i, true); err != nil {
			return err
		}
	}
	for i, name := range defaultAccounts {
		account := model.Account{BookID: bookID, Name: name, SortOrder: i, IsSystem: true}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&account).Error; err != nil {
			return err
		}
	}
	member := model.BookMember{BookID: bookID, UserID: userID, Role: "owner"}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&member).Error
}

func createCategoryIfMissing(db *gorm.DB, bookID uint, typ, name string, order int, system bool) error {
	category := model.Category{BookID: bookID, Type: typ, Name: name, SortOrder: order, IsSystem: system}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&category).Error
}

func (s *AccountingService) Bootstrap(userID uint, month string) (Bootstrap, error) {
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return Bootstrap{}, err
	}
	categories, err := s.Categories(userID)
	if err != nil {
		return Bootstrap{}, err
	}
	accounts, err := s.Accounts(userID)
	if err != nil {
		return Bootstrap{}, err
	}
	budget, err := s.Budget(userID, month)
	if err != nil {
		return Bootstrap{}, err
	}
	return Bootstrap{Book: book, Categories: categories, Accounts: accounts, MonthlyBudgetFen: budget.AmountFen}, nil
}

func (s *AccountingService) Categories(userID uint) (map[string][]string, error) {
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return nil, err
	}
	var rows []model.Category
	if err := s.db.Where("book_id = ?", book.ID).Order("type asc, sort_order asc, id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := map[string][]string{"expense": {}, "income": {}}
	for _, row := range rows {
		result[row.Type] = append(result[row.Type], row.Name)
	}
	return result, nil
}

func (s *AccountingService) Accounts(userID uint) ([]string, error) {
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return nil, err
	}
	var rows []model.Account
	if err := s.db.Where("book_id = ?", book.ID).Order("sort_order asc, id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]string, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.Name)
	}
	return items, nil
}

func (s *AccountingService) AddCategory(userID uint, input NameInput) (model.Category, error) {
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return model.Category{}, err
	}
	typ := normalizeType(input.Type)
	name := cleanName(input.Name)
	if name == "" {
		return model.Category{}, errors.New("name is required")
	}
	category := model.Category{BookID: book.ID, Type: typ, Name: name, SortOrder: 1000, IsSystem: false}
	err = s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&category).Error
	if err != nil {
		return model.Category{}, err
	}
	if category.ID == 0 {
		_ = s.db.Where("book_id = ? AND type = ? AND name = ?", book.ID, typ, name).First(&category).Error
	}
	return category, nil
}

func (s *AccountingService) AddAccount(userID uint, input NameInput) (model.Account, error) {
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return model.Account{}, err
	}
	name := cleanName(input.Name)
	if name == "" {
		return model.Account{}, errors.New("name is required")
	}
	account := model.Account{BookID: book.ID, Name: name, SortOrder: 1000, IsSystem: false}
	err = s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&account).Error
	if err != nil {
		return model.Account{}, err
	}
	if account.ID == 0 {
		_ = s.db.Where("book_id = ? AND name = ?", book.ID, name).First(&account).Error
	}
	return account, nil
}

func (s *AccountingService) Budget(userID uint, month string) (model.Budget, error) {
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return model.Budget{}, err
	}
	if !monthPattern.MatchString(month) {
		return model.Budget{BookID: book.ID, Month: month, AmountFen: 0}, nil
	}
	var budget model.Budget
	err = s.db.Where("book_id = ? AND month = ?", book.ID, month).First(&budget).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Budget{BookID: book.ID, Month: month, AmountFen: 0}, nil
	}
	return budget, err
}

func (s *AccountingService) SetBudget(userID uint, month string, input BudgetInput) (model.Budget, error) {
	if !monthPattern.MatchString(month) {
		return model.Budget{}, errors.New("invalid month")
	}
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return model.Budget{}, err
	}
	if input.AmountFen < 0 {
		input.AmountFen = 0
	}
	budget := model.Budget{BookID: book.ID, Month: month, AmountFen: input.AmountFen}
	err = s.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "book_id"}, {Name: "month"}}, DoUpdates: clause.AssignmentColumns([]string{"amount_fen", "updated_at"})}).Create(&budget).Error
	if err != nil {
		return model.Budget{}, err
	}
	_ = s.db.Where("book_id = ? AND month = ?", book.ID, month).First(&budget).Error
	return budget, nil
}

func (s *AccountingService) Transactions(userID uint, month string) ([]model.Transaction, error) {
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return nil, err
	}
	query := s.db.Where("book_id = ?", book.ID)
	if month != "" {
		if !monthPattern.MatchString(month) {
			return nil, errors.New("invalid month")
		}
		query = query.Where("tx_date >= ? AND tx_date < ?", month+"-01", nextMonth(month)+"-01")
	}
	var rows []model.Transaction
	err = query.Order("tx_date desc, created_at desc, id desc").Limit(1000).Find(&rows).Error
	return rows, err
}

func (s *AccountingService) CreateTransaction(userID uint, input TransactionInput) (model.Transaction, error) {
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return model.Transaction{}, err
	}
	if input.ClientID == "" {
		input.ClientID = fmt.Sprintf("server_%d", nowMillis())
	}
	var existing model.Transaction
	if err := s.db.Where("book_id = ? AND client_id = ?", book.ID, input.ClientID).First(&existing).Error; err == nil {
		return existing, nil
	}
	category, account, err := s.ensureCategoryAndAccount(book.ID, input)
	if err != nil {
		return model.Transaction{}, err
	}
	tx := model.Transaction{BookID: book.ID, ClientID: input.ClientID, Type: normalizeType(input.Type), AmountFen: input.AmountFen, CategoryID: category.ID, CategoryName: category.Name, AccountID: account.ID, AccountName: account.Name, TxDate: input.Date, Note: strings.TrimSpace(input.Note), Source: normalizeSource(input.Source)}
	if err := validateTransaction(tx); err != nil {
		return model.Transaction{}, err
	}
	err = s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&tx).Error
	if err != nil {
		return model.Transaction{}, err
	}
	if tx.ID == 0 {
		_ = s.db.Where("book_id = ? AND client_id = ?", book.ID, input.ClientID).First(&tx).Error
	}
	return tx, nil
}

func (s *AccountingService) UpdateTransaction(userID uint, id uint, input TransactionInput) (model.Transaction, error) {
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return model.Transaction{}, err
	}
	var tx model.Transaction
	if err := s.db.Where("id = ? AND book_id = ?", id, book.ID).First(&tx).Error; err != nil {
		return model.Transaction{}, err
	}
	category, account, err := s.ensureCategoryAndAccount(book.ID, input)
	if err != nil {
		return model.Transaction{}, err
	}
	tx.Type = normalizeType(input.Type)
	tx.AmountFen = input.AmountFen
	tx.CategoryID = category.ID
	tx.CategoryName = category.Name
	tx.AccountID = account.ID
	tx.AccountName = account.Name
	tx.TxDate = input.Date
	tx.Note = strings.TrimSpace(input.Note)
	if input.Source != "" {
		tx.Source = normalizeSource(input.Source)
	}
	if err := validateTransaction(tx); err != nil {
		return model.Transaction{}, err
	}
	return tx, s.db.Save(&tx).Error
}

func (s *AccountingService) DeleteTransaction(userID uint, id uint) error {
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return err
	}
	return s.db.Where("id = ? AND book_id = ?", id, book.ID).Delete(&model.Transaction{}).Error
}

func (s *AccountingService) Summary(userID uint, month string) (Summary, error) {
	if !monthPattern.MatchString(month) {
		return Summary{}, errors.New("invalid month")
	}
	book, err := ensureDefaultBook(s.db, userID)
	if err != nil {
		return Summary{}, err
	}
	var rows []model.Transaction
	if err := s.db.Where("book_id = ? AND tx_date >= ? AND tx_date < ?", book.ID, month+"-01", nextMonth(month)+"-01").Find(&rows).Error; err != nil {
		return Summary{}, err
	}
	summary := Summary{}
	expenseMap := map[string]int64{}
	incomeMap := map[string]int64{}
	for _, row := range rows {
		summary.Count++
		if row.Type == "income" {
			summary.Income += row.AmountFen
			incomeMap[row.CategoryName] += row.AmountFen
		} else {
			summary.Expense += row.AmountFen
			expenseMap[row.CategoryName] += row.AmountFen
		}
	}
	summary.Balance = summary.Income - summary.Expense
	budget, err := s.Budget(userID, month)
	if err != nil {
		return Summary{}, err
	}
	summary.Budget = budget.AmountFen
	if budget.AmountFen > 0 {
		summary.BudgetLeft = budget.AmountFen - summary.Expense
		pct := summary.Expense * 100 / budget.AmountFen
		if pct > 100 {
			pct = 100
		}
		summary.BudgetPercent = int(pct)
	}
	summary.ExpenseRanks = rank(expenseMap)
	summary.IncomeRanks = rank(incomeMap)
	return summary, nil
}

func (s *AccountingService) ensureCategoryAndAccount(bookID uint, input TransactionInput) (model.Category, model.Account, error) {
	categoryName := cleanName(input.Category)
	accountName := cleanName(input.Account)
	if categoryName == "" {
		categoryName = "其他"
	}
	if accountName == "" {
		accountName = "默认账户"
	}
	category := model.Category{BookID: bookID, Type: normalizeType(input.Type), Name: categoryName, SortOrder: 1000}
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&category).Error; err != nil {
		return category, model.Account{}, err
	}
	if category.ID == 0 {
		_ = s.db.Where("book_id = ? AND type = ? AND name = ?", bookID, category.Type, category.Name).First(&category).Error
	}
	account := model.Account{BookID: bookID, Name: accountName, SortOrder: 1000}
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&account).Error; err != nil {
		return category, account, err
	}
	if account.ID == 0 {
		_ = s.db.Where("book_id = ? AND name = ?", bookID, account.Name).First(&account).Error
	}
	return category, account, nil
}

func validateTransaction(tx model.Transaction) error {
	if tx.Type != "income" && tx.Type != "expense" {
		return errors.New("invalid type")
	}
	if tx.AmountFen <= 0 {
		return errors.New("amount must be positive")
	}
	if !datePattern.MatchString(tx.TxDate) {
		return errors.New("invalid date")
	}
	return nil
}

func normalizeType(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "income") {
		return "income"
	}
	return "expense"
}

func normalizeSource(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "manual"
	}
	if len(value) > 40 {
		return value[:40]
	}
	return value
}

func cleanName(value string) string {
	value = strings.TrimSpace(value)
	if len([]rune(value)) > 40 {
		return string([]rune(value)[:40])
	}
	return value
}

func nextMonth(month string) string {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return month
	}
	return t.AddDate(0, 1, 0).Format("2006-01")
}

func nowMillis() int64 { return time.Now().UnixNano() / int64(time.Millisecond) }

func rank(items map[string]int64) []CategoryRank {
	result := make([]CategoryRank, 0, len(items))
	for name, amount := range items {
		result = append(result, CategoryRank{Name: name, AmountFen: amount})
	}
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].AmountFen > result[i].AmountFen {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}
