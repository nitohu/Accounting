package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

/*
Query:

SELECT
    t.id,
    t.name,
    t.active,
    t.transaction_date,
    t.create_date,
    t.last_update,
    t.amount,
    t.transaction_type,
    u.name,
    a.name
FROM transactions AS t
JOIN users AS u ON t.user_id=u.id
JOIN accounts AS a ON t.account_id=a.id
WHERE t.user_id= *id*;
*/

// Transaction model
// TODO: Replace TransactionDate with better naming
// TODO: Implement Forecasted and Booked in database
type Transaction struct {
	// Database fields
	ID                int64
	Name              string
	Active            bool
	TransactionDate   time.Time
	CreateDate        time.Time
	LastUpdate        time.Time
	Amount            float64
	FromAccount       int64
	ToAccount         int64
	TransactionType   string
	Forecasted        bool
	ForecastedReverse bool
	Booked            bool
	BookedReverse     bool
	UserID            int

	// Computed fields
	FromAccountName    string
	ToAccountName      string
	TransactionDateStr string
}

// EmptyTransaction ..
func EmptyTransaction() Transaction {
	t := Transaction{
		ID:                0,
		Name:              "",
		Active:            false,
		TransactionDate:   time.Now().Local(),
		CreateDate:        time.Now().Local(),
		LastUpdate:        time.Now().Local(),
		Amount:            0.0,
		FromAccount:       0,
		ToAccount:         0,
		TransactionType:   "",
		Booked:            false,
		BookedReverse:     false,
		Forecasted:        false,
		ForecastedReverse: false,
		UserID:            0,
	}

	return t
}

func bookIntoAccount(cr *sql.DB, id int64, t *Transaction, invert bool) error {
	acc, err := FindAccountByID(cr, id)

	if err != nil {
		fmt.Println("Origin: bookIntoAccount()")
		fmt.Println("Hint: Finding Account")
		return err
	}

	err = acc.Book(cr, t, invert)

	if err != nil {
		fmt.Println("Origin: bookIntoAccount()")
		fmt.Println("Hint: Booking into Account")
		return err
	}

	return nil
}

// Create 's a transaction with the current values of the object
func (t *Transaction) Create(cr *sql.DB) error {

	fmt.Println("Transcation.Create function")

	// Requirements for creating a transaction
	if t.ID != 0 {
		return errors.New("This object already has a user id")
	} else if t.UserID == 0 {
		return errors.New("No user is linked to the transaction")
	} else if t.Amount == 0.0 {
		return errors.New("The Amount of this transaction is 0")
	}

	var id int64

	// Initializing variables
	query := "INSERT INTO transactions ( name, active, transaction_date, last_update, create_date, amount,"
	query += " account_id, to_account, transaction_type, user_id, booked, forecasted, booked_reverse,"
	query += " forecasted_reverse) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'f', 'f', 'f', 'f') RETURNING id;"

	t.CreateDate = time.Now().Local()
	t.LastUpdate = time.Now().Local()

	var err error

	// 4 cases of executing the query
	if t.ToAccount == 0 && t.FromAccount > 0 {
		err = cr.QueryRow(query,
			t.Name,
			t.Active,
			t.TransactionDate,
			t.LastUpdate,
			t.CreateDate,
			t.Amount,
			t.FromAccount,
			nil,
			t.TransactionType,
			t.UserID,
		).Scan(&id)
	} else if t.ToAccount > 0 && t.FromAccount > 0 {
		err = cr.QueryRow(query,
			t.Name,
			t.Active,
			t.TransactionDate,
			t.LastUpdate,
			t.CreateDate,
			t.Amount,
			t.FromAccount,
			t.ToAccount,
			t.TransactionType,
			t.UserID,
		).Scan(&id)
	} else if t.ToAccount == 0 && t.FromAccount == 0 {
		err = cr.QueryRow(query,
			t.Name,
			t.Active,
			t.TransactionDate,
			t.LastUpdate,
			t.CreateDate,
			t.Amount,
			nil,
			nil,
			t.TransactionType,
			t.UserID,
		).Scan(&id)
	} else if t.ToAccount > 0 && t.FromAccount == 0 {
		err = cr.QueryRow(query,
			t.Name,
			t.Active,
			t.TransactionDate,
			t.LastUpdate,
			t.CreateDate,
			t.Amount,
			nil,
			t.ToAccount,
			t.TransactionType,
			t.UserID,
		).Scan(&id)
	}

	if err != nil {
		fmt.Println("Origin: Transaction.Create")
		return err
	}

	// Writing id to object
	t.ID = id

	fmt.Printf("Received id by db: %d\n", id)
	fmt.Printf("Objects ID: %d\n", t.ID)

	fmt.Printf("FromAccount: %d\n", t.FromAccount)
	fmt.Printf("ToAccount: %d\n", t.ToAccount)

	// Book transaction into FromAccount if it's given
	if t.FromAccount > 0 {
		fromAccount, err := FindAccountByID(cr, t.FromAccount)
		fmt.Printf("FromAccount given, ID: %d\n", fromAccount.ID)

		if err != nil {
			fmt.Println("Origin: Transaction.Create")
			fmt.Println("Hint: if t.FromAccount > 0 == true")
			return err
		}

		err = fromAccount.Book(cr, t, true)

		if err != nil {
			fmt.Println("Origin: Transaction.Create")
			fmt.Println("Hint: if t.FromAccount > 0 == true")
			return err
		}

		err = t.Save(cr)

		if err != nil {
			fmt.Println("Origin: Transaction.Create")
			fmt.Println("Hint: if t.FromAccount > 0 == true, error saving transaction")
			return err
		}
	}

	// Book the transaction into ToAccount if it's given
	if t.ToAccount > 0 {
		toAccount, err := FindAccountByID(cr, t.ToAccount)
		fmt.Printf("ToAccount given, ID: %d\n", toAccount.ID)

		if err != nil {
			fmt.Println("Origin: Transaction.Create")
			fmt.Println("Hint: if t.ToAccount > 0 == true")
			return err
		}

		err = toAccount.Book(cr, t, false)

		if err != nil {
			fmt.Println("Origin: Transaction.Create")
			fmt.Println("Hint: if t.ToAccount > 0 == true")
			return err
		}

		err = t.Save(cr)

		if err != nil {
			fmt.Println("Origin: Transaction.Create")
			fmt.Println("Hint: if t.ToAccount > 0 == true, error saving transaction")
			return err
		}
	}

	return nil
}

// Save 's the current values of the object to the database
func (t *Transaction) Save(cr *sql.DB) error {
	fmt.Println("Transaction.Save function")

	if t.ID == 0 {
		return errors.New("This account as no ID, maybe create it first?")
	} else if t.Amount == 0.0 {
		return errors.New("The Amount of the transaction with the id " + fmt.Sprintf("%d", t.ID) + " is 0")
	}

	// Get old data
	var oldAmount float64
	var accountID, toAccountID interface{}
	var TransactionDateStr time.Time

	query := "SELECT amount, transaction_date, account_id, to_account FROM transactions WHERE id=$1"

	row := cr.QueryRow(query, t.ID)

	err := row.Scan(&oldAmount, &TransactionDateStr, &accountID, &toAccountID)

	if err != nil {
		fmt.Println("Origin: Transaction.Save()")
		fmt.Println("Hint: Receiving the old data")
		return err
	}

	// Write values to database
	query = "UPDATE transactions SET name=$2, active=$3, transaction_date=$4, last_update=$5, amount=$6, account_id=$7,"
	query += "to_account=$8, transaction_type=$9 WHERE id=$1"

	t.TransactionDate = time.Now().Local()

	var res sql.Result

	if t.ToAccount == 0 && t.FromAccount > 0 {
		res, err = cr.Exec(query,
			t.ID,
			t.Name,
			t.Active,
			t.TransactionDate,
			t.LastUpdate,
			t.Amount,
			t.FromAccount,
			nil,
			t.TransactionType,
		)
	} else if t.ToAccount > 0 && t.FromAccount > 0 {
		res, err = cr.Exec(query,
			t.ID,
			t.Name,
			t.Active,
			t.TransactionDate,
			t.LastUpdate,
			t.Amount,
			t.FromAccount,
			t.ToAccount,
			t.TransactionType,
		)
	} else if t.ToAccount == 0 && t.FromAccount == 0 {
		// TODO: This case shouldn't be allowed
		res, err = cr.Exec(query,
			t.ID,
			t.Name,
			t.Active,
			t.TransactionDate,
			t.LastUpdate,
			t.Amount,
			nil,
			nil,
			t.TransactionType,
		)
	} else if t.ToAccount > 0 && t.FromAccount == 0 {
		res, err = cr.Exec(query,
			t.ID,
			t.Name,
			t.Active,
			t.TransactionDate,
			t.LastUpdate,
			t.Amount,
			nil,
			t.ToAccount,
			t.TransactionType,
		)
	}

	if err != nil {
		fmt.Println("Origin: Transaction.Save()")
		fmt.Println("Hint: Writing the new data")
		return err
	}

	_, err = res.RowsAffected()

	if err != nil {
		fmt.Println("Origin: Transaction.Save()")
		fmt.Println("Hint: Getting number of affected rows")
		return err
	}

	/*
		Change balance on accounts
	*/

	// Check if the new amount if different from the old one
	// book into accounts if that is the case
	if t.Amount != oldAmount {
		tempTransaction := EmptyTransaction()
		// Calculating the difference
		tempTransaction.Amount += t.Amount - oldAmount

		// Book difference into old origin account
		if accountID != nil && accountID.(int64) == t.FromAccount {
			err = bookIntoAccount(cr, accountID.(int64), &tempTransaction, true)

			if err != nil {
				fmt.Println("Origin: Transaction.Save()")
				fmt.Println("Hint: The amount of the transaction changed")
				fmt.Println("Hint: Error while booking into old origin account")
				return err
			}
		}

		// Book difference into old receiving account
		if toAccountID != nil && toAccountID.(int64) == t.ToAccount {
			err = bookIntoAccount(cr, toAccountID.(int64), &tempTransaction, false)

			if err != nil {
				fmt.Println("Origin: Transaction.Save()")
				fmt.Println("Hint: The amount of the transaction changed")
				fmt.Println("Hint: Error while booking into old receiving account")
				return err
			}
		}
	}

	// Check if the origin account changed
	// and make sure there was an origin account before
	if accountID != t.FromAccount && accountID != nil {
		err := bookIntoAccount(cr, accountID.(int64), t, false)

		if err != nil {
			fmt.Println("Origin: Transaction.Save()")
			fmt.Println("Hint: Redo booking from old origin account")
			return err
		}

		err = bookIntoAccount(cr, t.FromAccount, t, false)

		if err != nil {
			fmt.Println("Origin: Transaction.Save()")
			fmt.Println("Hint: Book to new origin account")
			return err
		}

	}

	// Check if the receiving account changed
	// make also sure there was a receiving account before
	if toAccountID != t.ToAccount && toAccountID != nil {
		err := bookIntoAccount(cr, toAccountID.(int64), t, true)

		if err != nil {
			fmt.Println("Origin: Transaction.Save()")
			fmt.Println("Hint: Removing transaction from the old receiving account")
			return err
		}

		err = bookIntoAccount(cr, t.ToAccount, t, false)

		if err != nil {
			fmt.Println("Origin: Transaction.Save()")
			fmt.Println("Hint: Book to new origin account")
			return err
		}
	}

	// TODO: Implement change for TransactionDateStr

	if err != nil {
		return err
	}

	return nil
}

// Delete 's the transtaction
func (t *Transaction) Delete(cr *sql.DB) error {
	if t.ID == 0 {
		return errors.New("The account you want to delete does not have an id")
	}

	query := "DELETE FROM transactions WHERE id=$1"

	_, err := cr.Exec(query, t.ID)

	if err != nil {
		return err
	}

	return nil
}

// ComputeFields computes the fields which are not directly received
// from the database
func (t *Transaction) ComputeFields(cr *sql.DB) error {
	// Compute: FromAccountName
	if t.FromAccount != 0 {
		fromAccount, err := FindAccountByID(cr, t.FromAccount)

		if err != nil {
			return err
		}

		t.FromAccountName = fromAccount.Name
	} else {
		t.FromAccountName = "External Account"
	}

	// Compute: ToAccountName
	if t.ToAccount != 0 {
		toAccount, err := FindAccountByID(cr, t.ToAccount)

		if err != nil {
			return err
		}

		t.ToAccountName = toAccount.Name
	} else {
		t.ToAccountName = "External Account"
	}

	// Compute: TransactionDateStr
	t.TransactionDateStr = t.TransactionDate.String()

	return nil
}

// FindByID finds a transaction with it's id
func (t *Transaction) FindByID(cr *sql.DB, transactionID int64) error {
	query := "SELECT id, name, active, transaction_date, last_update, create_date, "
	query += "amount, account_id, to_account, transaction_type, user_id "
	query += "FROM transactions WHERE id=$1 "
	// query += "forecasted, booked_reverse, forecasted_reverse FROM transactions WHERE id=$1 "
	query += "ORDER BY transaction_date"

	var fromAccountID, toAccountID interface{}

	err := cr.QueryRow(query, transactionID).Scan(
		&t.ID,
		&t.Name,
		&t.Active,
		&t.TransactionDate,
		&t.LastUpdate,
		&t.CreateDate,
		&t.Amount,
		&fromAccountID,
		&toAccountID,
		&t.TransactionType,
		&t.UserID,
		// &t.Booked,
		// &t.Forecasted,
		// &t.BookedReverse,
		// &t.ForecastedReverse,
	)

	if fromAccountID != nil {
		t.FromAccount = fromAccountID.(int64)
	}

	if toAccountID != nil {
		t.ToAccount = toAccountID.(int64)
	}

	if err != nil {
		return err
	}

	t.ComputeFields(cr)

	if err != nil {
		return err
	}

	return nil
}

// FindTransactionByID is similar to FindByID but returns the transaction
func FindTransactionByID(cr *sql.DB, transactionID int64) (Transaction, error) {
	t := EmptyTransaction()

	err := t.FindByID(cr, transactionID)

	if err != nil {
		return t, err
	}

	return t, nil
}
