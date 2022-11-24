package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl"
	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
)

func main() {
	parseOption()

	f, err := os.Open(option.Config)
	if err != nil {
		errExit(err)
	}

	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		errExit(err)
	}
	if err := f.Close(); err != nil {
		errExit(err)
	}

	ctx := commonsCtx.WithUserID(context.Background(), "ADMIN")

	dm := parseDataManager(ctx, cfg.Postgre)
	defer dm.Close()

	switch cfg.Action {
	case "":
		fallthrough
	case create:
		createAccounts(ctx, dm, cfg)
	case reset:
		resetPassword(ctx, dm, cfg.IDs)
	case delete:
		deleteAccounts(ctx, dm, cfg.IDs)
	default:
		errExitWithDM(fmt.Errorf("unknown action"), dm)
	}

	fmt.Println("Operation completed!")
	printDragon()
}

type dBConnection struct {
	Name     string `yaml:"name"`
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
	Schema   string `yaml:"schema"`
}

const (
	create = "create"
	reset  = "reset"
	delete = "delete"
)

type config struct {
	Postgre dBConnection `yaml:"postgres"`
	Action  string       `yaml:"action"`
	IDs     []string     `yaml:"ids"`
}

var option struct {
	Config string `short:"c" long:"config" description:"Configuration file" required:"true"`
}

func createAccounts(ctx context.Context, dm mcom.DataManager, cfg config) {
	req := parseRequest(cfg.IDs)

	if err := dm.CreateAccounts(ctx, req); err != nil {
		errExitWithDM(err, dm)
	}
}

func resetPassword(ctx context.Context, dm mcom.DataManager, ids []string) {
	eg := errgroup.Group{}

	for _, id := range ids {
		id := id
		eg.Go(func() error {
			if err := dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
				UserID: id,
			}, mcom.ResetPassword()); err != nil {
				printErr(id, err)
				return err
			}
			fmt.Println("user_id: " + id + " password reset successfully")
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		if err := dm.Close(); err != nil {
			errExit(err)
		}
		os.Exit(1)
	}
}

func deleteAccounts(ctx context.Context, dm mcom.DataManager, ids []string) {
	eg := errgroup.Group{}

	for _, id := range ids {
		id := id
		eg.Go(func() error {
			if err := dm.DeleteAccount(ctx, mcom.DeleteAccountRequest{
				ID: id,
			}); err != nil {
				printErr(id, err)
				return err
			}
			fmt.Println("user_id: " + id + " has been deleted successfully")
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		if err := dm.Close(); err != nil {
			errExit(err)
		}
		os.Exit(1)
	}
}

type commandLineOptionsGroup struct {
	ShortDescription string
	LongDescription  string
	Options          interface{}
}

func setOption() []commandLineOptionsGroup {
	return []commandLineOptionsGroup{{
		ShortDescription: "configuration setting",
		LongDescription:  "configuration setting for user authorization",
		Options:          &option,
	}}
}

func parseOption() {
	options := setOption()

	parser := flags.NewParser(nil, flags.Default)
	for _, o := range options {
		if _, err := parser.AddGroup(o.ShortDescription, o.LongDescription, o.Options); err != nil {
			errExit(err)
		}
	}
	if _, err := parser.Parse(); err != nil {
		errExit(err)
	}
}

func printDragon() {
	fmt.Println(`                                                    __----~~~~~~~~~~~------___`)
	fmt.Println(`                                   .  .   ~~//====......          __--~ ~~    `)
	fmt.Println(`                   -.            \_|//     |||\\  ~~~~~~::::... /~            `)
	fmt.Println(`                ___-==_       _-~o~  \/    |||  \\            _/~~-           `)
	fmt.Println(`        __---~~~.==~||\=_    -_--~/_-~|-   |\\   \\        _/~                `)
	fmt.Println(`    _-~~     .=~    |  \\-_    '-~7  /-   /  ||    \      /                   `)
	fmt.Println(`  .~       .~       |   \\ -_    /  /-   /   ||      \   /                    `)
	fmt.Println(` /  ____  /         |     \\ ~-_/  /|- _/   .||       \ /                     `)
	fmt.Println(` |~~    ~~|--~~~~--_ \     ~==-/   | \~--===~~        .\                      `)
	fmt.Println(`          '         ~-|      /|    |-~\~~       __--~~                        `)
	fmt.Println(`                      |-~~-_/ |    |   ~\_   _-~            /\                `)
	fmt.Println(`                           /  \     \__   \/~                \__              `)
	fmt.Println(`                       _--~ _/ | .-~~____--~-/                  ~~==.         `)
	fmt.Println(`                      ((->/~   '.|||' -_|    ~~-/ ,              . _||        `)
	fmt.Println(`                                 -_     ~\      ~~---l__i__i__i--~~_/         `)
	fmt.Println(`                                 _-~-__   ~)  \--______________--~~           `)
	fmt.Println(`                               //.-~~~-~_--~- |-------~~~~~~~~                `)
	fmt.Println(`                                      //.-~~~--\)                             `)
}

func parseDataManager(ctx context.Context, dbConnection dBConnection) mcom.DataManager {
	dm, err := impl.New(ctx, impl.PGConfig{
		Address:  dbConnection.Address,
		Port:     dbConnection.Port,
		UserName: dbConnection.UserName,
		Password: dbConnection.Password,
		Database: dbConnection.Name,
	}, impl.WithPostgreSQLSchema(dbConnection.Schema))
	if err != nil {
		errExit(err)
	}
	return dm
}

func parseRequest(ids []string) mcom.CreateAccountsRequest {
	req := make(mcom.CreateAccountsRequest, len(ids))

	for i, id := range ids {
		req[i] = mcom.CreateAccountRequest{
			ID:    id,
			Roles: []roles.Role{roles.Role_ROLE_UNSPECIFIED},
		}.WithDefaultPassword()
	}

	return req
}

func errExit(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func errExitWithDM(err error, dm mcom.DataManager) {
	fmt.Println(err.Error())
	if e := dm.Close(); e != nil {
		err = e
		errExit(err)
	}
	os.Exit(1)
}

func printErr(id string, err error) {
	fmt.Println("user_id: " + id + ", error:" + err.Error())
}
