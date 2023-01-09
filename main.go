package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
  "net/http"

	"database/sql"
	"encoding/gob"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/go-sql-driver/mysql"
)

// HTMLからリクエスト来た時にGo内でそのデータが受け取れるようにこのStructを用意する。
type Bookmark struct {
	Name    string `form:"bookName"`
	URL     string `form:"bookUrl"`
	Comment string `form:"bookcomment"`
}

type BookmarkJson struct {
	Name    string `json:"bookname"`
	URL     string `json:"URL"`
	Comment string `json:"Comment"`
}

// booklistテーブルと同じ構造。
type Record struct {
	ID       int
	Bookname string
	URL      string
	Comment  string
	Time     string
}

// loginで使用
type User struct {
	ID        string
	Username  string
	Email     string
	pswdHash  string
	CreatedAt string
	Active    string
	verHash   string
	timeout   string
}


//******* login機能 Userモデルの宣言********//
var db *sql.DB

var store = sessions.NewCookieStore([]byte("super-secret"))

func init() {
	store.Options.HttpOnly = true // since we are not accessing any cookies w/ JavaScript, set to true
	store.Options.Secure = true   // requires secuire HTTPS connection
	gob.Register(&User{})
}


func main() {
	// まずはデータベースに接続する。(パスワードは各々異なる)
	dsn := "host=localhost user=postgres password=Hach8686 dbname=test port=5432 sslmode=disable TimeZone=Asia/Tokyo"
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// エラーでたらプロセス終了
		log.Fatalf("Some error occured. Err: %s", err)
	}

	/*
	 * APIサーバーの設定をする。
	 * rはrouterの略で何のAPIを用意するかを定義する。
	 * postpage　GET、/showpage　GET、/user　POST
	 */

	
	// 最初に定義するやつ
	r := gin.Default()

	// ginに対して、使うHTMLのテンプレートがどこに置いてあるかを知らせる。
	r.LoadHTMLGlob("temp/*")

	// 用意していないエンドポイント以外を叩かれたら内部で/showpage　GETを叩いてデフォルトページを表示する様にする。
	r.NoRoute(func(c *gin.Context) {
		location := url.URL{Path: "/showpage"}
		c.Redirect(http.StatusFound, location.RequestURI())
	})



	//************* signup/login 機能開始 **************//  

	r.LoadHTMLGlob("temp/*.html")
	var err error

	// ?? 3306に繋ぐのはなぜ？？//
	db, err = sql.Open("mysql", "root:super-secret-password@tcp(localhost:3306)/gin_db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	authRouter := r.Group("/user", auth)

	r.GET("/", indexHandler)
	r.GET("/login", loginGEThandler)
	r.POST("/login", loginPOSThandler)

	authRouter.GET("/profile", profileHandler)

	err = r.Run("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}


// auth middleware
func auth(c *gin.Context) {
	fmt.Println("auth middleware running")
	session, _ := store.Get(c.Request, "session")
	fmt.Println("session:", session)
	_, ok := session.Values["user"]
	if !ok {
		c.HTML(http.StatusForbidden, "login.html", nil)
		c.Abort()
		return
	}
	fmt.Println("middleware done")
	c.Next()
}

// index page
func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

// loginGEThandler displays form for login
func loginGEThandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

// loginPOSThandler verifies login credentials
func loginPOSThandler(c *gin.Context) {
	var user User
	user.Username = c.PostForm("username")
	password := c.PostForm("password")
	err := user.getUserByUsername()
	if err != nil {
		fmt.Println("error selecting pswd_hash in db by Username, err:", err)
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"message": "check username and password"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.pswdHash), []byte(password))
	fmt.Println("err from bycrypt:", err)
	if err == nil {
		session, _ := store.Get(c.Request, "session")
		// session struct has field Values map[interface{}]interface{}
		session.Values["user"] = user
		// save before writing to response/return from handler
		session.Save(c.Request, c.Writer)
		c.HTML(http.StatusOK, "loggedin.html", gin.H{"username": user.Username})
		return
	}
	c.HTML(http.StatusUnauthorized, "login.html", gin.H{"message": "check username and password"})
}

// profileHandler displays profile information
func profileHandler(c *gin.Context) {
	session, _ := store.Get(c.Request, "session")
	var user = &User{}
	val := session.Values["user"]
	var ok bool
	if user, ok = val.(*User); !ok {
		fmt.Println("was not of type *User")
		c.HTML(http.StatusForbidden, "login.html", nil)
		return
	}
	c.HTML(http.StatusOK, "profile.html", gin.H{"user": user})
}

func (u *User) getUserByUsername() error {
	stmt := "SELECT * FROM users WHERE username = ?"
	row := db.QueryRow(stmt, u.Username)
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.pswdHash, &u.CreatedAt, &u.Active, &u.verHash, &u.timeout)
	if err != nil {
		fmt.Println("getUser() error selecting User, err:", err)
		return err
	}
	return nil


  // ************login 機能終了***********************//




	// POST用のページ（post.html）を返す。
	// c.HTMLというのはこのAPIのレスポンスとしてHTMLファイルを返すよ、という意味
	r.GET("/postpage", func(c *gin.Context) {
		c.HTML(http.StatusOK, "post.html", gin.H{})
	})

	// 結果を表示するページを返す。
	r.GET("/showpage", func(c *gin.Context) {
		var records []Record
		// &recordsをDBに渡して、取得したデータを割り付ける。
		dbc := conn.Raw("SELECT id, bookname,url,comment,to_char(time,'YYYY-MM-DD HH24:MI:SS') AS time FROM booklist ORDER BY id").Scan(&records)
		if dbc.Error != nil {
			fmt.Print(dbc.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// レスポンスとして、show.htmlを返すが、一緒にrecordsも返している。これにより、HTML内でデータをマッピング表示することができる。
		c.HTML(http.StatusOK, "show.html", gin.H{
			"Books": records,
		})
	})

	// データを登録するAPI。POST用のページ（post.html）の内部で送信ボタンを押すと呼ばれるAPI。
	r.POST("/book", func(c *gin.Context) {

		var book Bookmark
		if err := c.ShouldBind(&book); err != nil {
			fmt.Print(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid argument"})
			return
		}
		var record Record
		// 以下の様にしてInsert文を書いて、リクエストデータをDBに書くこむ。.Scan(&record)はDBに書き込む際に必要らしい。
		// recordはbooklistテーブルと構造を同じにしている。(Gormのお作法)
		dbc := conn.Raw(
			"insert into booklist(bookname, url, comment) values(?, ?, ?)",
			book.Name, book.URL, book.Comment).Scan(&record)
		if dbc.Error != nil {
			fmt.Print(dbc.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// DBへの保存が成功したら結果を表示するページに戻るために/showpageのAPIを内部で読んでそちらでページの表示を行う。
		location := url.URL{Path: "/showpage"}
		c.Redirect(http.StatusFound, location.RequestURI())
	})

	// PUT 内容のupdate
	r.PUT("/bookupdate/:id", func(c *gin.Context) {
		id := c.Param("id")
		var book BookmarkJson
		if err := c.ShouldBind(&book); err != nil {
			fmt.Print(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid argument"})
			return
		}
		var record Record
		dbc := conn.Raw(
			"UPDATE booklist SET bookname=?, url=?, comment=? where id=?",
			book.Name, book.URL, book.Comment, id).Scan(&record)
		if dbc.Error != nil {
			fmt.Print(dbc.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

	})

	// データの削除
	r.DELETE("/book/:id", func(c *gin.Context) {
		id := c.Param("id")
		fmt.Println("id is ", id)
		var records []Record
		dbc := conn.Raw("DELETE FROM booklist where id=?", id).Scan(&records)

		if dbc.Error != nil {
			fmt.Print(dbc.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	})

	// showpageで書籍名をinputしてボタン押したら→入力した書籍名とおなじ列をdeleteする
	r.DELETE("/book/select/:bookname", func(c *gin.Context) {
		bookname := c.Param("bookname")
		fmt.Println("bookname is ", bookname)

		// レコードが存在するか確認。
		var record Record
		dbc := conn.Raw("SELECT * FROM booklist where bookname=?", bookname).Scan(&record)
		if dbc.Error != nil {
			fmt.Print(dbc.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		// レコードがなければNotFoundエラーを返す
		if dbc.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{})
			return
		}

		var records []Record
		dbc = conn.Raw("DELETE FROM booklist where bookname=?", bookname).Scan(&records)

		if dbc.Error != nil {
			fmt.Print(dbc.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		c.JSON(http.StatusNoContent, gin.H{})
	})

	// GET APIでid(ここを押せるようにする)を押すと、そのデータだけが表示されたページに遷移する
	// 結果を表示するページを返す。

	r.GET("/book/transition/:id", func(c *gin.Context) {
		id := c.Param("id")
		fmt.Println("id is ", id)
		// c.HTML(http.StatusOK, "select.html", gin.H{"id": id})
		var records []Record
		dbc := conn.Raw("SELECT id, bookname,url,comment,to_char(time,'YYYY-MM-DD HH24:MI:SS') AS time FROM booklist where id=?", id).Scan(&records)

		if dbc.Error != nil {
			fmt.Print(dbc.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.HTML(http.StatusOK, "select.html", gin.H{
			"Selects": records[0],
		})
	})

	// サーバーを立ち上げた瞬間は一旦ここまで実行されてListening状態となる。
	// r.POST( や　r.GET(　等の関数はAPIが呼ばれる度に実行される。
	r.Run()

}
}
