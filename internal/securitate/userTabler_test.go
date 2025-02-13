package securitate

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ctx context.Context

func TestDBstruct_AddUser(t *testing.T) {
	type args struct {
		userName string
		password string
	}
	tests := []struct {
		name      string
		args      args
		isErr     bool
		errString string
	}{
		{
			name: "Nice adding1",
			args: args{
				userName: "us1",
				password: "pass1",
			},
			isErr: false,
		},
		{
			name: "Nice adding2",
			args: args{
				userName: "us2",
				password: "pass2",
			},
			isErr: false,
		},
		{
			name: "Nice adding3",
			args: args{
				userName: "us3",
				password: "pass3",
			},
			isErr: false,
		},
		{
			name: "Duplicate adding",
			args: args{
				userName: "us3",
				password: "pass1",
			},
			isErr:     true,
			errString: "23505",
		},
	}

	
	err := dropTables()
	if err != nil {
		fmt.Printf("drop tables  %v", err)
		return
	}

	dataBase, err := ConnectToDB(ctx)
	if err != nil {
		fmt.Printf("database connection error  %v", err)
		return
	}
	defer dataBase.DB.Close(ctx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dataBase.AddUser(ctx, tt.args.userName, tt.args.password, "TOKENa")
			assert.Equal(t, tt.isErr, err != nil)
			if err != nil {
				assert.ErrorContains(t, err, tt.errString)
			}
		})
	}
	tt := tests[0]
	t.Run("correct password", func(t *testing.T) {
		err := dataBase.CheckUserPassword(ctx, tt.args.userName, tt.args.password)
		assert.Equal(t, tt.isErr, err != nil)
		if err != nil {
			assert.ErrorContains(t, err, tt.errString)
		}
	})

	//	tt = tests[0]
	t.Run("wrong password", func(t *testing.T) {
		err := dataBase.CheckUserPassword(ctx, tt.args.userName, tt.args.password+"a")
		assert.Equal(t, tt.isErr, err == nil)
		if err != nil {
			assert.ErrorContains(t, err, "password not match")
		}
	})
	t.Run("Right User", func(t *testing.T) {
		err := dataBase.IfUserExists(ctx, tt.args.userName)
		assert.Equal(t, tt.isErr, err != nil)
		if err != nil {
			assert.ErrorContains(t, err, "QueryRow, error is")
		}
	})
	t.Run("Wrong User", func(t *testing.T) {
		err := dataBase.IfUserExists(ctx, tt.args.userName+"a")
		assert.Equal(t, tt.isErr, err == nil)
		if err != nil {
			assert.ErrorContains(t, err, "QueryRow, error is")
		}
	})

}
func TestDBstruct_AddOrder(t *testing.T) {
	type args struct {
		userName    string
		orderNumber int64
		tokenStr    string
	}
	tests := []struct {
		name     string
		args     args
		noErr    bool
		errOrder string
		errToken string
	}{
		{
			name: "Nice Order",
			args: args{
				userName:    "us1",
				orderNumber: 1234,
				tokenStr:    "userexists",
			},
			noErr: true,
		},
		{
			name: "No user",
			args: args{
				userName:    "us2222",
				orderNumber: 3456,
				tokenStr:    "no user",
			},
			noErr:    false,
			errOrder: "add ORDER",
			errToken: "add TOKEN",
		},
	}
	ctx = context.Background()
	dataBase, err := ConnectToDB(ctx)
	if err != nil {
		fmt.Printf("database connection error  %v", err)
		return
	}
	defer dataBase.DB.Close(ctx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dataBase.AddOrder(ctx, tt.args.userName, tt.args.orderNumber)
			assert.Equal(t, tt.noErr, err == nil)
			if err != nil {
				assert.ErrorContains(t, err, tt.errOrder)
			}
			// err = dataBase.UpdateToken(ctx, tt.args.userName, tt.args.tokenStr)
			// assert.Equal(t, tt.noErr, err == nil)
			// if err != nil {
			// 	assert.ErrorContains(t, err, tt.errToken)
			// }
		})
	}

}

func dropTables() error {
	ctx = context.Background()
	dataBase, err := ConnectToDB(ctx)
	if err != nil {
		return fmt.Errorf("database connection error  %w", err)
	}
	for _, tab := range []string{"orders", "tokens", "withdrawn", "accounts"} {
		dropOrder := "DROP TABLE " + tab + " ;"
		tag, err := dataBase.DB.Exec(ctx, dropOrder)
		if err != nil {
			return fmt.Errorf("error DROP users table. Tag is \"%s\" error is %w", tag.String(), err)
		}
	}
	dataBase.DB.Close(ctx)
	return nil
}
