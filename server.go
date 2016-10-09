package main

import (
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
	"github.com/iris-contrib/template/django"
	"github.com/kataras/iris"
)

const SESSION_KEY string = "user_id"

var DBConn *dbr.Connection

type ToDoRow struct {
	ID          int
	Title       string
	Description string
	CreatedOn   string
}

func ensureNotLoggedIn(c *iris.Context) {
	if c.Session().GetString(SESSION_KEY) != "" {
		c.Redirect("/")
	} else {
		c.Next()
	}
}

func ensureLoggedIn(c *iris.Context) {
	if c.Session().GetString(SESSION_KEY) != "" {
		c.Next()
	} else {
		c.Redirect("/login")
	}
}

func DB() *dbr.Session {
	return DBConn.NewSession(nil)
}

func main() {

	var err error
	DBConn, err = dbr.Open("mysql", "user:password@/TO_DO_LIST?charset=utf8", nil)

	if err != nil {
		panic(err)
	}

	iris.UseTemplate(django.New()).Directory("./template", ".html")

	iris.Config.IsDevelopment = true // reloads the templates on each request, defaults to false

	iris.Static("/static", "./static", 1)

	public := iris.Party("/", ensureNotLoggedIn)

	public.Get("login", func(c *iris.Context) {
		c.MustRender("login.html", map[string]interface{}{
			"title": "Login",
			"flash": c.GetFlashes(),
		})
	})

	public.Post("login", func(c *iris.Context) {

		var id string
		err := DB().Select("id").
			From("user").
			Where("email=? AND password=?", c.FormValueString("email"), c.FormValueString("password")).
			LoadValue(&id)

		if err != nil {
			c.SetFlash("error", "Invalid login credentials")
			c.Redirect("/login")
		} else {
			c.Session().Set(SESSION_KEY, id)
			c.Redirect("/")
		}

	})

	public.Get("signup", func(c *iris.Context) {
		c.MustRender("signup.html", map[string]interface{}{
			"title": "Signup",
		})
	})

	public.Post("signup", func(c *iris.Context) {

		name := c.FormValueString("name")
		email := c.FormValueString("email")
		password := c.FormValueString("password")

		result, err := DB().InsertInto("user").Columns("name", "email", "password").Values(name, email, password).Exec()

		if err != nil {
			c.Panic()
		}

		id, _ := result.LastInsertId()

		c.Session().Set(SESSION_KEY, strconv.FormatInt(id, 10))

		c.Redirect("/")

	})

	private := iris.Party("/", ensureLoggedIn)

	private.Get("", func(c *iris.Context) {

		var data []ToDoRow
		var name string

		DB().Select("id", "title", "description", "created_on").
			From("to_do_item").
			Where("user_id=?", c.Session().GetString(SESSION_KEY)).
			LoadStructs(&data)

		DB().Select("name").
			From("user").
			Where("id=?", c.Session().GetString(SESSION_KEY)).
			LoadValue(&name)

		c.MustRender("index.html", map[string]interface{}{
			"title": "Your To-Do List",
			"flash": c.GetFlashes(),
			"ToDos": data,
			"name":  name,
		})

	})

	private.Post("create", func(c *iris.Context) {

		title := c.FormValueString("title")
		description := c.FormValueString("description")

		_, err := DB().InsertInto("to_do_item").
			Columns("user_id", "title", "description", "created_on").
			Values(
				c.Session().GetString(SESSION_KEY),
				title,
				description,
				dbr.Expr("NOW()"),
			).
			Exec()

		if err != nil {
			c.Log("%v", err)
			c.SetFlash("error", "There was a problem creating your to-do item!")
		} else {
			c.SetFlash("success", "to-do item created!")
		}

		c.Redirect("/")

	})

	private.Post("update/:id", func(c *iris.Context) {

		id, err := c.ParamInt("id")

		if err != nil {
			c.SetFlash("error", "Attempting to update invalid to-do item")
			c.Redirect("/")
		}

		title := c.FormValueString("title")
		description := c.FormValueString("description")

		_, err = DB().Update("to_do_item").
			Set("title", title).
			Set("description", description).
			Where("id=?", id).
			Where("user_id=?", c.Session().GetString(SESSION_KEY)).
			Exec()

		if err != nil {
			c.Log("%v", err)
			c.SetFlash("error", "There was a problem updating your to-do item!")
		} else {
			c.SetFlash("success", "to-do item updated!")
		}

		c.Redirect("/")

	})

	private.Post("delete/:id", func(c *iris.Context) {

		id, err := c.ParamInt("id")

		if err != nil {
			c.SetFlash("error", "Attempting to delete invalid to-do item")
			c.Redirect("/")
		}

		_, err = DB().DeleteFrom("to_do_item").
			Where("id=?", id).
			Where("user_id=?", c.Session().GetString(SESSION_KEY)).
			Exec()

		if err != nil {
			c.SetFlash("error", "There was a problem deleting your to-do item!")
		} else {
			c.SetFlash("success", "to-do item deleted!")
		}

		c.Redirect("/")

	})

	private.Post("complete/:id", func(c *iris.Context) {

		id, err := c.ParamInt("id")

		if err != nil {
			c.SetFlash("error", "Attempting to complete invalid to-do item")
			c.Redirect("/")
		}

		_, err = DB().Update("to_do_item").
			Set("completed_on=NOW()", nil).
			Where("id=?", id).
			Where("user_id=?", c.Session().GetString(SESSION_KEY)).
			Exec()

		if err != nil {
			c.SetFlash("error", "There was a problem marking your to-do item as complete!")
		} else {
			c.SetFlash("success", "to-do item completed!")
		}

		c.Redirect("/")

	})

	private.Get("logout", func(c *iris.Context) {
		if c.Session().GetString(SESSION_KEY) != "" {
			c.SessionDestroy()
		}
		c.Redirect("/")
	})

	iris.Get("/test", func(c *iris.Context) {
		c.MustRender("test.html", map[string]interface{}{
			"title":    "Test Page Title",
			"var_name": "bobby",
		})
	})

	iris.Listen(":8080")

}
