package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

//******* login機能 Userモデルの宣言********//
type User struct {
	Company  string `form:"company" binding:"required"`
	Username string `form:"username" binding:"required" gorm:"unique;not null"`
	password string `form:"password" binding:"required"`
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

	r := gin.Default()

	// ginに対して、使うHTMLのテンプレートがどこに置いてあるかを知らせる。
	r.LoadHTMLGlob("temp/*")

	// 用意していないエンドポイント以外を叩かれたら内部で/showpage　GETを叩いてデフォルトページを表示する様にする。
	r.NoRoute(func(c *gin.Context) {
		location := url.URL{Path: "/showpage"}
		c.Redirect(http.StatusFound, location.RequestURI())
	})


	//************* signup/login 機能 **************//  
	// (途中・・・)

	// r.GET("/signup", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "sign.html", gin.H{})
	// })

	// // ユーザー登録
	// r.POST("/signup", func(c *gin.Context) {
	// 	var form User
	// 	// バリデーション処理
	// 	if err := c.Bind(&form); err != nil {
	// 		c.HTML(http.StatusBadRequest, "signup.html", gin.H{"err": err})
	// 		c.Abort()
	// 	} else {
	// 		username := c.PostForm("username")
	// 		password := c.PostForm("password")
	// 		// 登録ユーザーが重複していた場合にはじく処理
	// 		if err := createUser(username, password); err != nil {
	// 			c.HTML(http.StatusBadRequest, "signup.html", gin.H{"err": err})
	// 		}
	// 		c.Redirect(302, "/")
	// 	}
	// })	

	// // ユーザーログイン画面
	
	// r.GET("/login", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "login.html", gin.H{})
	// })

	// r.POST("/login", func(c *gin.Context) {

	// 	// DBから取得したユーザーパスワード(Hash)
	// 	dbPassword := getUser(c.PostForm("username")).Password
	// 	log.Println(dbPassword)
	// 	// フォームから取得したユーザーパスワード
	// 	formPassword := c.PostForm("password")

	// 	// ユーザーパスワードの比較
	// 	if err := crypto.CompareHashAndPassword(dbPassword, formPassword); err != nil {
	// 			log.Println("ログインできませんでした")
	// 			c.HTML(http.StatusBadRequest, "login.html", gin.H{"err": err})
	// 			c.Abort()
	// 	} else {
	// 			log.Println("ログインできました")
	// 			c.Redirect(302, "/")
	// 	}

		// ユーザーを一件取得
		// func getUser(username string) User {
		// db := gormConnect()
		// var user User
		// db.First(&user, "username = ?", username)
		// db.Close()
		// return user
		// }

	// })
	
	//************* signup/login 機能 **************//  





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

// ログイン
