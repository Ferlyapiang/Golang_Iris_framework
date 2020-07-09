package controller

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/kataras/iris/sessions"
	uuid "github.com/satori/go.uuid"
	"github.com/wuriyanto48/replacer"

	"socmed/src/modules/profile/model"
	"socmed/src/modules/profile/usecase"
)

//ProfileIDKey
const ProfileIDKey = "ProfileID"

//ProfileController
type ProfileController struct {
	Ctx iris.Context

	Session *sessions.Session

	ProfileUsecase usecase.ProfileUsecase
}

//getCurrentProfileID
func (c *ProfileController) getCurrentProfileID() string {
	return c.Session.GetString(ProfileIDKey)
}

//isProfileLoggend
func (c *ProfileController) isProfileLoggedIn() bool {
	return c.getCurrentProfileID() != ""
}

func (c *ProfileController) loggout() {
	c.Session.Destroy()
}

//localHost:3000/profile/register GET
func (c *ProfileController) GetRegister() mvc.Result {
	if c.isProfileLoggedIn() {
		c.loggout()
	}

	return mvc.View{
		Name: "profile/register.html",
		Data: iris.Map{"Title": "Profile Registration"},
	}
}

//localHost:3000/profile/register POST
//PostRegister
func (c *ProfileController) PostRegister() mvc.Result {

	firstName := c.Ctx.FormValue("fist_name")
	lastName := c.Ctx.FormValue("last_name")
	email := c.Ctx.FormValue("email")
	password := c.Ctx.FormValue("password")

	if firstName == "" || lastName == "" || email == "" || password == "" {
		return mvc.Response{
			Path: "/profile/register",
		}
	}

	id := uuid.NewV4()

	profileImage, err := c.uploadImage(c.Ctx, id.String())

	if err != nil {
		return mvc.Response{
			Path: "/profile/register",
		}
	}

	var profile model.Profile
	profile.ID = id.String()
	profile.FirstName = firstName
	profile.LastName = lastName
	profile.Email = email
	profile.Password = password
	profile.ImageProfile = profileImage
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()

	_, err = c.ProfileUsecase.SaveProfile(&profile)

	if err != nil {
		return mvc.Response{
			Path: "/profile/register",
		}
	}

	c.Session.Set(ProfileIDKey, profile.ID)

	return mvc.Response{
		Path: "/profile/me",
	}
}

//localHost:3000/profile/register GET
//GetLogin
func (c *ProfileController) GetLogin() mvc.Result {
	if c.isProfileLoggedIn() {
		c.loggout()
	}

	return mvc.View{
		Name: "profile/login.html",
		Data: iris.Map{"Title": "Login"},
	}
}

//PostLogin
func (c *ProfileController) PostLogin() mvc.Result {

	email := c.Ctx.FormValue("email")
	password := c.Ctx.FormValue("password")

	if email == "" || password == "" {
		return mvc.Response{
			Path: "/profile/login",
		}
	}

	profile, err := c.ProfileUsecase.GetByEmail(email)

	if err != nil {
		return mvc.Response{
			Path: "/profile/login",
		}
	}

	if !profile.IsValidPassword(password) {
		return mvc.Response{
			Path: "/profile/login",
		}
	}

	c.Session.Set(ProfileIDKey, profile.ID)

	return mvc.Response{
		Path: "/profile/me",
	}
}

// localHost:3000/Profile/me GET
//GetMe
func (c *ProfileController) GetMe() mvc.Result {
	if !c.isProfileLoggedIn() {
		return mvc.Response{
			Path: "/profile/login",
		}
	}

	profile, err := c.ProfileUsecase.GetByID(c.getCurrentProfileID())

	if err != nil {
		c.loggout()
		c.GetMe()
	}

	return mvc.View{
		Name: "profile/me.html",
		Data: iris.Map{"Title": "My Profile", "Profile": profile},
	}
}

//LocalHost:3000/profile/logout
func (c *ProfileController) AnyLogout() {
	if c.isProfileLoggedIn() {
		c.loggout()
	}

	c.Ctx.Redirect("/profile/login")
}

//uploadImage
func (c *ProfileController) uploadImage(ctx iris.Context, id string) (string, error) {
	//get image from view

	file, info, err := ctx.FormFile("image_profile")

	if err != nil {
		return "", err
	}

	defer file.Close()

	fileName := fmt.Sprintf("%s%s%s", id, "_", replacer.Replace(info.Filename, "_"))

	out, err := os.OpenFile("./web/public/image/profile/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return "", err
	}
	defer out.Close()

	io.Copy(out, file)

	return fileName, nil
}
