package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/muhammadolammi/jobmatchapi/internal/auth"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"

	// "github.com/muhammadolammi/jobmatchapi/internal/helpers"
	"golang.org/x/crypto/bcrypt"
)

func (apiConfig *Config) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	body := struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		Role            string `json:"role"`
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		CompanyName     string `json:"company_name"`
		CompanyWebsite  string `json:"company_website"`
		CompanySize     int32  `json:"company_size"`
		CompanyIndustry string `json:"company_industry"`
	}{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)

	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error decoding request body. err: %v", err))
		return
	}
	// avoid admin sign up early
	if body.Role == "admin" {
		helpers.RespondWithError(w, http.StatusUnauthorized, "Unauthorized signup")
		return
	}
	if body.Email == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Enter a mail")
		return
	}
	if body.Password == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Enter a password")
		return
	}

	// check if user exist
	userExist, err := apiConfig.DB.UserExists(r.Context(), body.Email)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error validating user. err: %v", err))
		return
	}
	if userExist {
		helpers.RespondWithError(w, http.StatusBadRequest, "User already exist. Login")
		return
	}
	// Validate role and role in enum
	if body.Role == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Enter the user role.")
		return
	}
	if body.Role != "employer" && body.Role != "job_seeker" {
		helpers.RespondWithError(w, http.StatusBadRequest, "User  role must be one of (employer, job_seeker")
		return
	}
	//  Validate payload base on role.
	if body.Role == "employer" {
		if body.CompanyName == "" {
			helpers.RespondWithError(w, http.StatusBadRequest, "Enter the employer company_name")
			return
		}
		if body.CompanyIndustry == "" {
			helpers.RespondWithError(w, http.StatusBadRequest, "Enter the employer company_industry")
			return
		}
		if body.CompanySize == 0 {
			helpers.RespondWithError(w, http.StatusBadRequest, "Enter the employer company_size")
			return
		}
		if body.CompanyWebsite == "" {
			helpers.RespondWithError(w, http.StatusBadRequest, "Enter the employer company_website")
			return
		}

	}
	if body.Role == "job_seeker" {
		if body.FirstName == "" {
			helpers.RespondWithError(w, http.StatusBadRequest, "Enter the user first name")
			return
		}
		if body.LastName == "" {
			helpers.RespondWithError(w, http.StatusBadRequest, "Enter the user last name")
			return
		}

	}
	// { create the user}
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error hashing password. err: %v", err))
		return
	}
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))

	user, err := apiConfig.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:    body.Email,
		Password: string(hashedPassword),
		Role:     body.Role,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error creating user. err: %v", err))
		return
	}
	//  create detail base on role
	if user.Role == "employer" {
		_, err = apiConfig.DB.CreateEmployer(r.Context(), database.CreateEmployerParams{
			UserID:          user.ID,
			CompanyName:     body.CompanyName,
			CompanyWebsite:  body.CompanyWebsite,
			CompanySize:     body.CompanySize,
			CompanyIndustry: body.CompanyIndustry,
		})
		if err != nil {
			helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error creating employer details, kindly update your details. err: %v", err))
			return

		}

	}
	if user.Role == "job_seeker" {
		_, err = apiConfig.DB.CreateJobSeeker(r.Context(), database.CreateJobSeekerParams{
			UserID:    user.ID,
			LastName:  body.LastName,
			FirstName: body.FirstName,
		})
		if err != nil {
			helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error creating job seeker details, kindly update your details. err: %v", err))
			return
		}
	}
	helpers.RespondWithJson(w, 200, "signup successful")
}

func (apiConfig *Config) LoginHandler(w http.ResponseWriter, r *http.Request) {
	body := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)

	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error decoding body from http request. err: %v", err))
		return
	}
	if body.Email == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Enter a mail.")
		return
	}
	if body.Password == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "Enter a password.")
		return
	}
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))

	userExist, err := apiConfig.DB.UserExists(r.Context(), body.Email)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error validating user. err: %v", err))
		return
	}
	if !userExist {
		helpers.RespondWithError(w, http.StatusUnauthorized, "No User with this mail. Signup")
		return
	}

	user, err := apiConfig.DB.GetUserWithEmail(r.Context(), body.Email)

	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error getting user. err: %v", err))
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		if strings.Contains(err.Error(), `hashedPassword is not the hash of the given password`) {
			helpers.RespondWithError(w, http.StatusUnauthorized, "Wrong password.")
			return
		}
		helpers.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf(" err: %v", err))
		return
	}
	// create refresh token

	err = auth.CreateRefreshToken([]byte(apiConfig.JwtKey), user.ID, 24*7*6, w, apiConfig.DB)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error creating refresh token. err: %v", err))
		return
	}
	access_token, err := auth.MakeJwtTokenString([]byte(apiConfig.JwtKey), user.ID.String(), "access_token", 15)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error making jwt token. err: %v", err))
		return
	}
	response := struct {
		AccessToken string `json:"access_token"`
		// ExpiresAt   time.Time `json:"expires_at"`
	}{
		AccessToken: access_token,
	}

	helpers.RespondWithJson(w, 200, response)
}

func (apiConfig *Config) PasswordChangeHandler(w http.ResponseWriter, r *http.Request) {

	body := struct {
		Email       string `json:"email"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error decoding request body. err: %v", err))
		return
	}
	if body.Email == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Enter a mail. err: %v", err))
		return
	}
	if body.OldPassword == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Enter a password. err: %v", err))
		return
	}
	if body.NewPassword == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Enter a new password. err: %v", err))
		return
	}

	user, err := apiConfig.DB.GetUserWithEmail(r.Context(), body.Email)

	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error getting user. err: %v", err))
		return
	}
	// AUTHENTICATE THE USER
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.OldPassword))
	if err != nil {
		if strings.Contains(err.Error(), `hashedPassword is not the hash of the given password`) {
			helpers.RespondWithError(w, http.StatusUnauthorized, "Wrong password.")
			return
		}
		helpers.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf(" err: %v", err))
		return
	}
	// UPDATE THE PASSWORD
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 10)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error hashing password. err: %v", err))
		return
	}

	err = apiConfig.DB.UpdatePassword(r.Context(), database.UpdatePasswordParams{
		Email:    body.Email,
		Password: string(newHashedPassword),
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error updating password. err: %v", err))
		return
	}
	err = auth.UpdateRefreshToken([]byte(apiConfig.JwtKey), user.ID, 24*7*60, w, apiConfig.DB)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error updating refresh token. err: %v", err))
		return
	}
	helpers.RespondWithJson(w, 200, "Password Updated")
}

func (apiConfig *Config) GetUserHandler(w http.ResponseWriter, r *http.Request, user User) {

	helpers.RespondWithJson(w, 200, user)
}

func (apiConfig *Config) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	refreshtoken, err := r.Cookie("refresh_token")
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error getting refreshToken, Try login again. err: %v", err))
		return
	}

	refreshclaims := &jwt.RegisteredClaims{}

	_, err = jwt.ParseWithClaims(
		refreshtoken.Value,
		refreshclaims,
		func(token *jwt.Token) (interface{}, error) { return []byte(apiConfig.JwtKey), nil },
	)

	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error parsing jwt claims. err: %v", err))
		return
	}
	// Make sure refresh token exist in db
	refreshTokenExists, err := apiConfig.DB.RefreshTokenExists(context.Background(), refreshtoken.Value)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error checking refresh token in db. err: %v", err))
		return
	}
	if !refreshTokenExists {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("refresh token doesnt exist. err: %v", err))
		return

	}
	userIdString := refreshclaims.Issuer
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error parsing user id, err: %v", err))
		return
	}

	user, err := apiConfig.DB.GetUser(r.Context(), userId)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error getting user with id, err: %v", err))
		return
	}

	refreshExpiration := refreshclaims.ExpiresAt.Time

	if refreshExpiration.Before(time.Now().UTC()) {
		helpers.RespondWithError(w, http.StatusUnauthorized, "refresh token expired")
		return
	}

	access_token, err := auth.MakeJwtTokenString([]byte(apiConfig.JwtKey), user.ID.String(), "access_token", 15)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error making jwt token. err: %v", err))
		return
	}
	response := struct {
		AccessToken string `json:"access_token"`
		// ExpiresAt   time.Time `json:"expires_at"`
	}{
		AccessToken: access_token,
	}
	helpers.RespondWithJson(w, 200, response)

}

func (apiConfig *Config) Validate(w http.ResponseWriter, r *http.Request) {

	helpers.RespondWithJson(w, 200, "logged in")
}
