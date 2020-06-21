package store

//go:generate bash -c "sed 's/Foo/Project/g;s/foo/gtproject/g' foo.tpl > projects_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
)

type Project struct {
	Synced
	gtproject.Project
	Client   Client `gorm:"foreignkey:ClientPK"`
	ClientPK uint
}

func (obj *Project) BeforeCreate(db *gorm.DB) error {
	obj.loadClientFromToggl(db)
	return nil
}

func (obj *Project) BeforeSave(db *gorm.DB) error {
	obj.loadClientFromToggl(db)
	return nil
}

func (obj Project) GetClient() dto.Client {
	if obj.Client.PK == 0 {
		panic("not preloaded")
	}
	return obj.Client
}
func (obj Project) GetHexColor() string {
	return obj.HexColor
}

func (obj *Project) loadClientFromToggl(db *gorm.DB) {
	if obj.Client.PK == 0 {
		// Загружаем клиента из БД по внутреннему ключу
		if obj.ClientPK != 0 {
			if !MustOrNotFound(db.Where("pk = ?", obj.ClientPK).First(&obj.Client)) {
				obj.ClientPK = 0
			}
		}
	}
	if obj.Client.PK == 0 {
		// По внутреннему ключу клиент не найден, ищем через внешний ключ
		if MustOrNotFound(db.Where("id = ?", obj.CId).First(&obj.Client)) {
			// Нашли по внешнему ключу, проставляем внутренний
			obj.ClientPK = obj.Client.PK
		}
	}
}

type Projects []Project

type ProjectManager gorm.DB

func (m *ProjectManager) List() Projects {
	return listProjects(m.gorm().Order("name ASC"))
}

func (m *ProjectManager) Get(extId uint64) *Project {
	var project Project
	if MustOrNotFound(m.gorm().
		Where("id = ?", extId).First(&project)) {
		return &project
	}
	return nil
}
