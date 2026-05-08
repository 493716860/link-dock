package models

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/config"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Id               int64     `orm:"pk;auto;column(id)" json:"id,string"`
	Username         string    `orm:"column(username);unique" json:"username"`
	Password         string    `orm:"column(password)" json:"-"`
	SecurityQuestion string    `orm:"column(security_question);null" json:"securityQuestion"`
	SecurityAnswer   string    `orm:"column(security_answer);null" json:"-"`
	IsSuperAdmin     bool      `orm:"column(is_super_admin)" json:"is_super_admin"`
	CreatedAt        time.Time `orm:"auto_now_add;type(datetime);column(created_at)" json:"created_at"`
}

type Category struct {
	Id        int64     `orm:"pk;auto;column(id)" json:"id,string"`
	UserId    int64     `orm:"column(user_id)" json:"userId,string"`
	Name      string    `orm:"column(name)" json:"name"`
	IconName  string    `orm:"column(iconName)" json:"iconName"`
	SortOrder int       `orm:"column(sort_order);default(0)" json:"sortOrder"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
}

type Site struct {
	Id             int64      `orm:"pk;auto;column(id)" json:"id,string"`
	CategoryId     int64      `orm:"column(category_id)" json:"categoryId,string"`
	UserId         int64      `orm:"column(user_id)" json:"userId,string"`
	SortOrder      int        `orm:"column(sort_order);default(0)" json:"sortOrder"`
	Name           string     `orm:"column(name)" json:"name"`
	Description    string     `orm:"column(description);null" json:"description"`
	Url            string     `orm:"column(url)" json:"url"`
	TagsText       string     `orm:"column(tags_text);null" json:"tagsText"`
	IsFavorite     bool       `orm:"column(is_favorite);default(false)" json:"isFavorite"`
	WorkflowStatus string     `orm:"column(workflow_status);default(unorganized)" json:"workflowStatus"`
	VisitCount     int        `orm:"column(visit_count);default(0)" json:"visitCount"`
	LastVisitedAt  *time.Time `orm:"column(last_visited_at);null;type(datetime)" json:"lastVisitedAt,omitempty"`
	LogoText       string     `orm:"column(logoText);null" json:"logoText"`
	LogoColor      string     `orm:"column(logoColor);null" json:"logoColor"`
	LogoUrl        string     `orm:"column(logoUrl);null" json:"icon"`
	CreatedAt      time.Time  `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt      time.Time  `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
}

func init() {
	// 获取数据库类型，优先读取环境变量 DB_DRIVER，然后是配置文件中的 db_driver
	dbType := os.Getenv("DB_DRIVER")
	if dbType == "" {
		dbType, _ = config.String("db_driver")
	}
	if dbType == "" {
		dbType = "sqlite" // 默认使用 sqlite
	}

	fmt.Printf(">>> Initializing Database (Driver: %s)...\n", dbType)

	var err error
	if dbType == "mysql" {
		orm.RegisterDriver("mysql", orm.DRMySQL)

		dbUser, _ := config.String("db_user")
		dbPass, _ := config.String("db_pass")
		dbHost, _ := config.String("db_host")
		dbPort, _ := config.String("db_port")
		dbName, _ := config.String("db_name")

		// 允许通过环境变量覆盖具体连接参数
		if val := os.Getenv("DB_USER"); val != "" {
			dbUser = val
		}
		if val := os.Getenv("DB_PASS"); val != "" {
			dbPass = val
		}
		if val := os.Getenv("DB_HOST"); val != "" {
			dbHost = val
		}
		if val := os.Getenv("DB_PORT"); val != "" {
			dbPort = val
		}
		if val := os.Getenv("DB_NAME"); val != "" {
			dbName = val
		}

		dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			dbUser, dbPass, dbHost, dbPort, dbName)
		err = orm.RegisterDataBase("default", "mysql", dataSource)
	} else {
		// 默认为 sqlite
		orm.RegisterDriver("sqlite3", orm.DRSqlite)
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			// 如果没指定具体路径，也没从环境变量读取，尝试从 app.conf 读取（可选扩展）
			dbPath, _ = config.String("db_path")
		}
		if dbPath == "" {
			dbPath = "linkdock.db"
		}
		err = orm.RegisterDataBase("default", "sqlite3", dbPath)
	}

	if err != nil {
		fmt.Printf(">>> Critical Error registering database: %v\n", err)
	} else {
		fmt.Printf(">>> Database registered successfully using %s driver.\n", dbType)
	}

	// 注册 Model
	orm.RegisterModel(new(User), new(Category), new(Site))
}

func EnsureSiteSortOrderColumn() {
	o := orm.NewOrm()
	_, err := o.Raw("ALTER TABLE site ADD COLUMN sort_order integer DEFAULT 0").Exec()
	if err == nil {
		fmt.Println(">>> Added site.sort_order column.")
		return
	}

	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "duplicate column") || strings.Contains(errMsg, "already exists") {
		return
	}

	fmt.Printf(">>> Skipped site.sort_order migration: %v\n", err)
}

func EnsureUserRecoveryColumns() {
	o := orm.NewOrm()
	ensureColumn(o, "ALTER TABLE user ADD COLUMN security_question varchar(255)")
	ensureColumn(o, "ALTER TABLE user ADD COLUMN security_answer varchar(255)")
}

func EnsureSiteEnhancementColumns() {
	o := orm.NewOrm()
	ensureColumn(o, "ALTER TABLE site ADD COLUMN tags_text text")
	ensureColumn(o, "ALTER TABLE site ADD COLUMN is_favorite bool DEFAULT 0")
	ensureColumn(o, "ALTER TABLE site ADD COLUMN workflow_status varchar(32) DEFAULT 'unorganized'")
	ensureColumn(o, "ALTER TABLE site ADD COLUMN visit_count integer DEFAULT 0")
	ensureColumn(o, "ALTER TABLE site ADD COLUMN last_visited_at datetime")
}

func ensureColumn(o orm.Ormer, sql string) {
	_, err := o.Raw(sql).Exec()
	if err == nil {
		fmt.Printf(">>> Applied migration: %s\n", sql)
		return
	}

	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "duplicate column") || strings.Contains(errMsg, "already exists") {
		return
	}

	fmt.Printf(">>> Skipped migration (%s): %v\n", sql, err)
}
