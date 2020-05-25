package admin

import (
	"github.com/qor/admin"

	"github.com/lucmichalski/finance-contrib/business.financialpost.com/models"
)

const menuName = "business.financialpost.com"

// ConfigureAdmin configure admin interface
func ConfigureAdmin(Admin *admin.Admin) {

	Admin.AddMenu(&admin.Menu{Name: menuName, Priority: 1})

	// Add Setting page
	Admin.AddResource(&models.SettingBusinessFinancialPost{}, &admin.Config{
		Name:      menuName + " Settings",
		Menu:      []string{menuName},
		Singleton: true,
		Priority:  1,
	})

}
