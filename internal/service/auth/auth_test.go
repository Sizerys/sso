package auth

import (
	"errors"
	er "sso/internal/domain/errors"
	"sso/internal/domain/models"
	"sso/internal/service/mocks"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

type TestLoginStruct struct {
	name             string
	expErr           error
	pass             string
	mockUserSaver    func(mock *mocks.MockSaverUser)
	mockUserProvider func(mock *mocks.MockUserProvider)
	mockAppProvider  func(mock *mocks.MockAppProvider)
	mockRtProvider   func(mock *mocks.MockRefreshTokenProvider)
}

func TestLogin(t *testing.T) {
	tests := []TestLoginStruct{
		{
			name:   "success login",
			pass:   "testPassword",
			expErr: nil,
			mockUserProvider: func(mock *mocks.MockUserProvider) {
				mock.EXPECT().User(gomock.Any(), "testEmail").Return(models.User{
					ID:       1,
					Email:    "testEmail",
					PassHash: []byte("testPassword"),
				}, nil)
			},
			mockAppProvider: func(mock *mocks.MockAppProvider) {
				mock.EXPECT().App(gomock.Any(), int32(1)).Return(models.App{
					ID:     1,
					Name:   "testName",
					Secret: "testSecret",
				}, nil)
			},
			mockRtProvider: func(mock *mocks.MockRefreshTokenProvider) {
				mock.EXPECT().SaveRefreshToken(gomock.Any(), int32(1), gomock.Any(),
					time.Now().Add(224*time.Hour))
			},
		},
		{
			name:   "user not found",
			expErr: er.ErrUserNotFound,
		},
		{
			name:   "get user error",
			expErr: errors.New("Some error"),
		},
		{
			name:   "faile compare hash",
			pass:   "testPassword",
			expErr: er.ErrInvalidCredentials,
			mockUserProvider: func(mock *mocks.MockUserProvider) {
				mock.EXPECT().User(gomock.Any(), "testEmail").Return(models.User{
					ID:       1,
					Email:    "testEmail",
					PassHash: []byte("testAnotherPassword"),
				}, nil)
			},
		},
		{
			name:   "not found app",
			pass:   "testPassword",
			expErr: er.ErrAppNotFound,
			mockUserProvider: func(mock *mocks.MockUserProvider) {
				mock.EXPECT().User(gomock.Any(), "testEmail").Return(models.User{
					ID:       1,
					Email:    "testEmail",
					PassHash: []byte("testPassword"),
				}, nil)
			},
			mockAppProvider: func(mock *mocks.MockAppProvider) {
				mock.EXPECT().App(gomock.Any(), int32(1)).Return(models.App{}, er.ErrAppNotFound)
			},
		},
		{
			name: "get app error",
			pass: "testPassword",
			mockUserProvider: func(mock *mocks.MockUserProvider) {
				mock.EXPECT().User(gomock.Any(), "testEmail").Return(models.User{
					ID:       1,
					Email:    "testEmail",
					PassHash: []byte("testPassword"),
				}, nil)
			},
		},
		{
			name: "refresh token already exists",
			pass: "testPassword",
			mockUserProvider: func(mock *mocks.MockUserProvider) {
				mock.EXPECT().User(gomock.Any(), "testEmail").Return(models.User{
					ID:       1,
					Email:    "testEmail",
					PassHash: []byte("testPassword"),
				}, nil)
			},
			mockAppProvider: func(mock *mocks.MockAppProvider) {
				mock.EXPECT().App(gomock.Any(), int32(1)).Return(models.App{
					ID:     1,
					Name:   "testName",
					Secret: "testSecret",
				}, nil)
			},
			mockRtProvider: func(mock *mocks.MockRefreshTokenProvider) {
				mock.EXPECT().SaveRefreshToken(gomock.Any(), int32(1), gomock.Any(),
					time.Now().Add(224*time.Hour)).Return(er.ErrRefreshTokenAlreadyExists)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			defer ctrl.Finish()
		})
	}

}
