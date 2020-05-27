package admin

import (
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-contrib/seekingalpha.com/models"
)

const menuName = "seekingalpha.com"

// ConfigureAdmin configure admin interface
func ConfigureAdmin(Admin *admin.Admin) {

	Admin.AddMenu(&admin.Menu{Name: menuName, Priority: 1})

	// Add Setting page
	Admin.AddResource(&models.SettingSeekingAlpha{}, &admin.Config{
		Name:      menuName + " Settings",
		Menu:      []string{menuName},
		Singleton: true,
		Priority:  1,
	})

}
