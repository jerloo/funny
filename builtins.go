package funny

import (
	"context"
	"crypto/md5"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/guonaihong/gout"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/xerrors"
)

//go:embed builtins.funny
var BuiltinsDotFunny string

// BuiltinFunction function handler
type BuiltinFunction func(interpreter *Interpreter, args []Value) Value

var (
	// FUNCTIONS all builtin functions
	FUNCTIONS = map[string]BuiltinFunction{
		"echo":         Echo,
		"echoln":       Echoln,
		"now":          Now,
		"b64en":        Base64Encode,
		"b64de":        Base64Decode,
		"assert":       Assert,
		"len":          Len,
		"md5":          Md5,
		"max":          Max,
		"min":          Min,
		"typeof":       Typeof,
		"uuid":         UUID,
		"httpreq":      HttpRequest,
		"env":          Env,
		"strjoin":      StrJoin,
		"strsplit":     StrSplit,
		"str":          Str,
		"int":          Int,
		"jwten":        JwtEncode,
		"jwtde":        JwtDecode,
		"sqlquery":     SqlQuery,
		"sqlexec":      SqlExec,
		"sqlexecfile":  SqlExecFile,
		"format":       FormatData,
		"dumpruntimes": DumpRuntimes,
		"readtext":     ReadText,
		"writetext":    WriteText,
		"readjson":     ReadJson,
		"writejson":    WriteJson,
	}
)

// ackEq check function arguments count valid
func ackEq(args []Value, count int) {
	if len(args) != count {
		panic(fmt.Sprintf("%d arguments required but got %d", count, len(args)))
	}
}

// ackGt check function arguments count valid
func ackGt(args []Value, count int) {
	if len(args) <= count {
		panic(fmt.Sprintf("greater than %d arguments required but got %d", count, len(args)))
	}
}

// Echo builtin function echos one or every item in a array
func Echo(interpreter *Interpreter, args []Value) Value {
	for _, item := range args {
		switch v := item.(type) {
		case map[string]Value:
		case map[string]interface{}:
			bts, err := json.Marshal(&v)
			if err != nil {
				panic(err)
			}
			fmt.Print(string(bts))
		default:
			fmt.Print(item)
		}
	}
	return nil
}

// Echoln builtin function echos one or every item in a array
func Echoln(interpreter *Interpreter, args []Value) Value {
	for index, item := range args {
		switch v := item.(type) {
		case map[string]Value:
		case map[string]interface{}:
			bts, err := json.Marshal(&v)
			if err != nil {
				panic(err)
			}
			fmt.Print(string(bts))
		default:
			fmt.Print(item)
		}

		if index == len(args)-1 {
			fmt.Print("\n")
		}
	}
	return nil
}

// Now builtin function return now time
func Now(interpreter *Interpreter, args []Value) Value {
	return Value(time.Now())
}

// Base64Encode return base64 encoded string
func Base64Encode(interpreter *Interpreter, args []Value) Value {
	base64encode := func(val string) string {
		return base64.StdEncoding.EncodeToString([]byte(val))
	}
	if len(args) == 1 {
		return Value(base64encode(args[0].(string)))
	}
	var results []string
	for _, item := range args {
		results = append(results, base64encode(item.(string)))
	}
	return Value(results)
}

// Base64Decode return base64 decoded string
func Base64Decode(interpreter *Interpreter, args []Value) Value {
	base64decode := func(val string) string {
		sb, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			panic(err)
		}
		return string(sb)
	}
	if len(args) == 1 {
		return Value(base64decode(args[0].(string)))
	}
	var results []string
	for _, item := range args {
		results = append(results, base64decode(item.(string)))
	}
	return Value(results)
}

// Assert return the value that has been given
func Assert(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 1)
	if val, ok := args[0].(bool); ok {
		if val {
			return Value(args[0])
		}
		panic("assert false")
	}
	panic("assert type error, only support [bool]")
}

// Len return then length of the given list
func Len(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 1)
	switch v := args[0].(type) {
	case *List:
		return Value(len(v.Values))
	case string:
		return Value(len(v))
	case []interface{}:
		return Value(len(v))
	}
	panic(P(fmt.Sprintf("len type error, only support [list, string] %s", Typing(args[0])), interpreter.Current))
}

// Md5 return then length of the given list
func Md5(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 1)
	switch v := args[0].(type) {
	case string:
		md5Ctx := md5.New()
		md5Ctx.Write([]byte(v))
		return hex.EncodeToString(md5Ctx.Sum(nil))
	default:
		break
	}
	panic("md5 type error, only support [string]")
}

// Max return then length of the given list
func Max(interpreter *Interpreter, args []Value) Value {
	ackGt(args, 1)
	switch v := args[0].(type) {
	case int:
		flag := v
		for _, item := range args[1:] {
			if val, ok := item.(int); ok {
				if val > flag {
					flag = val
				}
			}
		}
		return Value(flag)
	case *List:
		flag := interpreter.EvalExpression(v.Values[0])
		if flagA, ok := flag.(int); ok {
			for _, item := range v.Values {
				val := interpreter.EvalExpression(item)
				if val, ok := val.(int); ok {
					if val > flagA {
						flagA = val
					}
				}
			}
			return Value(flagA)
		}
	default:
		break
	}
	panic("max type error, only support [int]")
}

// Min return then length of the given list
func Min(interpreter *Interpreter, args []Value) Value {
	ackGt(args, 1)
	switch v := args[0].(type) {
	case int:
		flag := v
		for _, item := range args[1:] {
			if val, ok := item.(int); ok {
				if val < flag {
					flag = val
				}
			}
		}
		return Value(flag)
	case *List:
		flag := interpreter.EvalExpression(v.Values[0])
		if flagA, ok := flag.(int); ok {
			for _, item := range v.Values {
				val := interpreter.EvalExpression(item)
				if val, ok := val.(int); ok {
					if val < flagA {
						flagA = val
					}
				}
			}
			return Value(flagA)
		}
	default:
		break
	}
	panic("min type error, only support [int]")
}

// Typeof builtin function echos one or every item in a array
func Typeof(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 1)
	return Typing(args[0])
}

// UUID builtin function return a uuid string value
func UUID(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 0)
	u1 := uuid.NewV4()
	return Value(u1)
}

// HttpRequest builtin function for http request
func HttpRequest(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 5)
	method := ""
	url := ""
	data := make(map[string]Value)
	headers := map[string]interface{}{
		"User-Agent": "Funny HttpRequest",
		"Accept":     "*/*",
	}
	debug := false
	if m, ok := args[0].(string); ok {
		method = m
	}
	if u, ok := args[1].(string); ok {
		url = u
	}
	if d, ok := args[2].(map[string]Value); ok {
		data = d
	}
	if h, ok := args[3].(map[string]Value); ok {
		for key, val := range h {
			headers[key] = val
		}
	}
	if de, ok := args[4].(bool); ok {
		debug = de
		if !debug {
			debug = interpreter.Debug()
		}
	}
	switch method {
	case "GET":
		jsonResult := make(map[string]interface{})
		err := gout.GET(url).Debug(debug).SetQuery(data).SetHeader(headers).BindJSON(&jsonResult).Do()
		if err != nil {
			panic(xerrors.Errorf("response not json format %w", err))
		}
		return Value(jsonResult)
	case "POST":
		jsonResult := make(map[string]interface{})
		err := gout.POST(url).Debug(debug).SetJSON(data).SetHeader(headers).BindJSON(&jsonResult).Do()
		if err != nil {
			panic(xerrors.Errorf("response not json format: %w", err))
		}
		return Value(jsonResult)
	case "PUT":
		jsonResult := make(map[string]interface{})
		err := gout.PUT(url).Debug(debug).SetHeader(headers).BindJSON(&jsonResult).Do()
		if err != nil {
			panic(xerrors.Errorf("response not json format: %w", err))
		}
		return Value(jsonResult)
	case "DELETE":
		jsonResult := make(map[string]interface{})
		err := gout.DELETE(url).Debug(debug).SetHeader(headers).BindJSON(&jsonResult).Do()
		if err != nil {
			panic(xerrors.Errorf("response not json format: %w", err))
		}
		return Value(jsonResult)
	}
	panic(fmt.Errorf("method %s not support yet", method))
}

// Env return the value of env key
func Env(interpreter *Interpreter, args []Value) Value {
	ackGt(args, 0)
	if key, ok := args[0].(string); ok {
		val := os.Getenv(key)
		if val == "" && len(args) > 1 {
			return Value(args[1])
		}
		return Value(val)
	}
	panic("env type error, env key only support [string]")
}

// StrJoin equal strings.Join
func StrJoin(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 2)
	if arr, ok := args[0].(*List); ok {
		var strArr []string
		for _, item := range arr.Values {
			val := interpreter.EvalExpression(item)
			strArr = append(strArr, fmt.Sprintf("%v", val))
		}
		if sp, o := args[1].(string); o {
			return strings.Join(strArr, sp)
		}
		panic("strjoin type error, join part only support [string]")
	}
	panic("strjoin type error, join data only support [array]")
}

// StrSplit equal strings.Split
func StrSplit(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 1)
	if key, ok := args[0].(string); ok {
		val := os.Getenv(key)
		if val == "" && len(args) > 1 {
			return Value(args[1])
		}
		return Value(val)
	}
	panic("strsplit type error, strsplit value only support [string]")
}

// Str like string(1)
func Str(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 1)
	return fmt.Sprint(args[0])
	// panic("str type error, str data only support [string]")
}

// Int like int('1')
func Int(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 1)
	if v, ok := args[0].(time.Time); ok {
		return Value(int(v.Unix()))
	}
	dataStr := fmt.Sprint(args[0])
	for _, ch := range dataStr {
		if !isNameStart(ch) {
			panic("int type error, int only support [int format]")
		}
	}
	panic("int type error, int only support [int format]")
}

// JwtEncode jwten(method, secret, claims) string
func JwtEncode(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 3)
	method := fmt.Sprint(args[0])
	secret := fmt.Sprint(args[1])

	if v, ok := args[2].(map[string]Value); ok {
		bts, err := json.Marshal(&v)
		if err != nil {
			panic(err)
		}
		var claims jwt.MapClaims
		err = json.Unmarshal(bts, &claims)
		if err != nil {
			panic(err)
		}
		m := jwt.SigningMethodHS256
		if method == "HS256" {
			m = jwt.SigningMethodHS256
		}
		token := jwt.NewWithClaims(m, claims)
		result, err := token.SignedString([]byte(secret))
		if err != nil {
			panic(err)
		}
		return Value(result)
	}
	panic("jwten type error, claims only support [map[string]interface{}]")
}

// JwtDecode jwtde(method, secret, token) string
func JwtDecode(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 3)
	// method := fmt.Sprint(args[0])
	secret := fmt.Sprint(args[1])
	tokenString := fmt.Sprint(args[2])
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		panic(err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid { // 校验token
		bts, err := json.Marshal(&claims)
		if err != nil {
			panic(err)
		}
		var result map[string]interface{}
		err = json.Unmarshal(bts, &result)
		if err != nil {
			panic(err)
		}
		return Value(result)
	}
	panic("jwtde type error, token not valid")
}

// SqlQuery sqlquery(connection, sqlRaw, args) string
func SqlQuery(interpreter *Interpreter, args []Value) Value {
	ackGt(args, 1)
	switch v := args[0].(type) {
	case map[string]Value:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			v["user"],
			v["password"],
			v["host"],
			v["port"],
			v["database"])

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		defer db.Close()
		var sqlArgs []interface{}
		for _, arg := range args[2:] {
			sqlArgs = append(sqlArgs, arg)
		}
		rows, err := db.Query(fmt.Sprint(args[1]), sqlArgs...)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		r := make([]map[string]interface{}, 0)
		for rows.Next() {
			cols, err := rows.Columns()
			if err != nil {
				panic(err)
			}
			fields := make([]interface{}, len(cols))
			for index := range fields {
				fields[index] = Value(new(interface{}))
			}
			err = rows.Scan(fields...)
			if err != nil {
				panic(err)
			}
			row := make(map[string]interface{})
			for index, col := range cols {
				row[col] = fields[index]
			}
			r = append(r, row)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			panic(err)
		}
		return Value(r)
	}
	panic("sqlquery type error, connection")
}

// SqlExec sqlexec(connection, sqlRaw, args) string
func SqlExec(interpreter *Interpreter, args []Value) Value {
	ackGt(args, 1)
	switch v := args[0].(type) {
	case map[string]Value:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			v["user"],
			v["password"],
			v["host"],
			v["port"],
			v["database"])

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		defer db.Close()
		var sqlArgs []interface{}
		for _, arg := range args[2:] {
			sqlArgs = append(sqlArgs, arg)
		}
		result, err := db.Exec(fmt.Sprint(args[1]), sqlArgs...)
		if err != nil {
			panic(err)
		}
		last, err := result.LastInsertId()
		if err != nil {
			panic(err)
		}
		row, err := result.RowsAffected()
		if err != nil {
			panic(err)
		}
		return Value(map[string]interface{}{
			"lastInsertId": last,
			"rowsAffected": row,
		})
	}
	panic("sqlexec type error, connection")
}

// FormatData format(data, formatStr) string
func FormatData(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 2)
	switch v := args[0].(type) {
	case time.Time:
		return Value(v.Format(args[1].(string)))
	}
	panic("format type error, data")
}

// DumpRuntimes dumpruntimes()
func DumpRuntimes(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 0)
	bts, err := json.MarshalIndent(&interpreter.Vars, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bts))
	return Value(string(bts))
}

// ReadText readtext()
func ReadText(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 1)
	if filename, fileOk := args[0].(string); fileOk {
		if !path.IsAbs(filename) {
			d := path.Dir(interpreter.Current.File)
			filename = path.Join(d, filename)
		}
		bts, err := os.ReadFile(filename)
		if err != nil {
			panic(xerrors.Errorf("read file error: %w", err))
		}
		return Value(string(bts))
	}
	panic("args type error")
}

// WriteText writetext(text)
func WriteText(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 2)
	if filename, fileOk := args[0].(string); fileOk {
		if text, textOk := args[1].(string); textOk {
			if !path.IsAbs(filename) {
				d := path.Dir(interpreter.Current.File)
				filename = path.Join(d, filename)
			}
			err := os.WriteFile(filename, []byte(text), fs.ModeAppend)
			if err != nil {
				panic(xerrors.Errorf("write error: %w", err))
			}
		}
	}
	panic("args type error")
}

// ReadJson readjson()
func ReadJson(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 1)
	if filename, fileOk := args[0].(string); fileOk {
		if !path.IsAbs(filename) {
			d := path.Dir(interpreter.Current.File)
			filename = path.Join(d, filename)
		}
		bts, err := os.ReadFile(filename)
		if err != nil {
			panic(xerrors.Errorf("read file error: %w", err))
		}
		var m map[string]Value
		err = json.Unmarshal(bts, &m)
		if err != nil {
			panic(xerrors.Errorf("read file error: %w", err))
		}
		return Value(m)
	}
	panic("args type error")
}

// WriteJson writejson(obj)
func WriteJson(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 2)
	if filename, fileOk := args[0].(string); fileOk {
		bts, err := json.Marshal(args[1])
		if err != nil {
			panic(err)
		}
		if !path.IsAbs(filename) {
			d := path.Dir(interpreter.Current.File)
			filename = path.Join(d, filename)
		}
		err = os.WriteFile(filename, []byte(bts), 0644)
		if err != nil {
			panic(xerrors.Errorf("write error: %w", err))
		}
		return Value(nil)
	}
	panic(P(fmt.Sprintf("args type error %s", Typing(args[0])), interpreter.Current))
}

// SqlExecFile sqlexecfile(connection, file)
func SqlExecFile(interpreter *Interpreter, args []Value) Value {
	ackEq(args, 2)
	if filename, fileOk := args[1].(string); fileOk {
		if !path.IsAbs(filename) {
			d := path.Dir(interpreter.Current.File)
			filename = path.Join(d, filename)
		}
		bts, err := os.ReadFile(filename)
		if err != nil {
			panic(xerrors.Errorf("read file error: %w", err))
		}
		v, connOk := args[0].(map[string]Value)
		if !connOk {
			panic(xerrors.Errorf("connection must dict"))
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
			v["user"],
			v["password"],
			v["host"],
			v["port"],
			v["database"])

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		defer db.Close()
		tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
		if err != nil {
			panic(err)
		}
		// sqls := strings.Split(string(bts), "\n")
		// for _, sql := range sqls {
		// 	_, err = tx.Exec(sql)
		// 	if err != nil {
		// 		tx.Rollback()
		// 		panic(err)
		// 	}
		// }
		_, err = tx.Exec(string(bts))
		if err != nil {
			tx.Rollback()
			panic(err)
		}
		tx.Commit()
		if err != nil {
			panic(err)
		}
		return Value(nil)
	}
	panic("args type error")
}
