// Lucas FOLLIOT
package todo

import (
	"context"
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Service struct {
	sessionStore SessionStore
	userStore    UserStore
	todoStore    TodoStore
}

type UpdateRequest struct {
	Text string
}

func NewService(redisURI string, pgURI string) *Service {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURI,
	})

	pgdb, err := pgxpool.Connect(context.Background(), pgURI)
	if err != nil {
		panic(err)
	}

	_ = rdb
	_ = pgdb
	// TODO: implement each interface
	return &Service{
		sessionStore: NewSessionStoreRedis(rdb),
		userStore:    NewUserStorePG(pgdb),
		todoStore:    NewTodoStorePG(pgdb),
	}
}

func (s *Service) SetupRoute(r gin.IRouter) {
	r.POST("/user/signup", s.Signup)
	r.GET("/user/:name", s.getUserByName)
	r.POST("/user/login", s.login)

	r.Use(s.EnsureLoggedIn())
	r.GET("/test", test)
	r.GET("/user/logout", s.logout)
	r.GET("/user", s.getUser)
	r.POST("/todos", s.createTodo)
	r.GET("/todos", s.getTodos)
	r.GET("/todos/:id", s.getTodo)
	r.PUT("/todos/:id", s.updateTodo)
	r.DELETE("/todos/:id", s.deleteTodo)

	// TODO:

	// Must be protected by a middleware checking the validity of the token in the cookie
	// 401 if not valid
}

func (s *Service) Signup(c *gin.Context) {
	var user User

	if err := c.BindJSON(&user); err != nil {
		return
	}

	u, err := s.userStore.Insert(user)
	if err != nil {
		c.IndentedJSON(http.StatusConflict, "User already exist")
	} else {
		token, _ := GenerateRandomString(20)
		userId := uuid.Must(uuid.FromString(u.Id))

		s.sessionStore.Add(userId, token)
		c.SetCookie("session_todo", token, 3600, "/", "localhost", false, false)
		c.IndentedJSON(http.StatusOK, u)
	}
}

func (s *Service) getUserByName(c *gin.Context) {
	name := c.Param("name")

	user, err := s.userStore.FindByName(name)
	if err != nil {
		c.IndentedJSON(http.StatusForbidden, "User not found")
	} else {
		c.IndentedJSON(http.StatusOK, user)
	}
}

func (s *Service) getUserByUuid(c *gin.Context) {
	name := c.Param("name")

	user, err := s.userStore.FindByName(name)
	if err != nil {
		panic(err)
	}

	c.IndentedJSON(http.StatusOK, user)
}

func (s *Service) login(c *gin.Context) {
	var user User

	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, "Bad params")
		return
	}

	token, _ := GenerateRandomString(20)
	u, err := s.userStore.FindByName(user.Name)
	if err != nil {
		c.IndentedJSON(http.StatusForbidden, "User not found")
		return
	}

	userId := uuid.Must(uuid.FromString(u.Id))

	s.sessionStore.Add(userId, token)
	c.SetCookie("session_todo", token, 3600, "/", "localhost", false, false)

	c.IndentedJSON(http.StatusOK, u)
}

func (s *Service) logout(c *gin.Context) {
	token, _ := c.Cookie("session_todo")
	s.sessionStore.Revoke(token)
	c.SetCookie("session_todo", "", 1, "/", "localhost", false, false)
	c.IndentedJSON(http.StatusOK, "Logout")
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func test(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "Test")
}

func (s *Service) EnsureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_todo")

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Error while parsing session cookie")
			return
		}

		userId, erro := s.sessionStore.FindByToken(token)

		if erro != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Any session where found")
			return
		}

		c.Set("userId", userId.String())

		c.Next()
	}
}

func (s *Service) getUser(c *gin.Context) {
	userId := c.GetString("userId")

	user, _ := s.userStore.FindById(uuid.FromStringOrNil(userId))

	c.IndentedJSON(http.StatusOK, user)
}

func (s *Service) createTodo(c *gin.Context) {
	var todo Todo
	userId := c.GetString("userId")

	if err := c.BindJSON(&todo); err != nil {
		return
	}

	todo.UserId = userId

	t, _ := s.todoStore.Add(todo)
	c.IndentedJSON(http.StatusOK, t)
}

func (s *Service) getTodos(c *gin.Context) {
	userId := c.GetString("userId")
	user, _ := s.userStore.FindById(uuid.FromStringOrNil(userId))

	todos, err := s.todoStore.FindByUserID(uuid.FromStringOrNil(user.Id))

	if err != nil {
		c.IndentedJSON(http.StatusForbidden, "Todo not found")
	} else {
		c.IndentedJSON(http.StatusOK, todos)
	}
}

func (s *Service) getTodo(c *gin.Context) {
	id := c.Param("id")

	todo, err := s.todoStore.FindByID(uuid.FromStringOrNil(id))

	if err != nil {
		c.IndentedJSON(http.StatusForbidden, "Todo not found")
	} else {
		c.IndentedJSON(http.StatusOK, todo)
	}
}

func (s *Service) updateTodo(c *gin.Context) {
	var updateRequest UpdateRequest
	id := c.Param("id")

	if err := c.BindJSON(&updateRequest); err != nil {
		panic(err)
	}

	todo, err := s.todoStore.UpdateText(uuid.FromStringOrNil(id), updateRequest.Text)

	if err != nil {
		c.IndentedJSON(http.StatusForbidden, "Todo not found")
	} else {
		c.IndentedJSON(http.StatusOK, todo)
	}
}

func (s *Service) deleteTodo(c *gin.Context) {
	id := c.Param("id")

	s.todoStore.Delete(uuid.FromStringOrNil(id))

	c.IndentedJSON(http.StatusAccepted, "Todo deleted")
}
