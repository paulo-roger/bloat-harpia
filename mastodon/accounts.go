package mastodon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"
)

type AccountPleroma struct {
	Relationship *Relationship `json:"relationship"`
}

// Account hold information for mastodon account.
type Account struct {
	ID             string          `json:"id"`
	Username       string          `json:"username"`
	Acct           string          `json:"acct"`
	DisplayName    string          `json:"display_name"`
	Locked         bool            `json:"locked"`
	CreatedAt      time.Time       `json:"created_at"`
	FollowersCount int64           `json:"followers_count"`
	FollowingCount int64           `json:"following_count"`
	StatusesCount  int64           `json:"statuses_count"`
	Note           string          `json:"note"`
	URL            string          `json:"url"`
	Avatar         string          `json:"avatar"`
	AvatarStatic   string          `json:"avatar_static"`
	Header         string          `json:"header"`
	HeaderStatic   string          `json:"header_static"`
	Emojis         []Emoji         `json:"emojis"`
	Moved          *Account        `json:"moved"`
	Fields         []Field         `json:"fields"`
	Bot            bool            `json:"bot"`
	Source         *AccountSource  `json:"source"`
	Pleroma        *AccountPleroma `json:"pleroma"`

	// Duplicate field for compatibilty with Pleroma
	FollowRequestsCount int64 `json:"follow_requests_count"`
}

// Field is a Mastodon account profile field.
type Field struct {
	Name       string    `json:"name"`
	Value      string    `json:"value"`
	VerifiedAt time.Time `json:"verified_at"`
}

// AccountSource is a Mastodon account profile field.
type AccountSource struct {
	Privacy             *string  `json:"privacy"`
	Sensitive           *bool    `json:"sensitive"`
	Language            *string  `json:"language"`
	Note                *string  `json:"note"`
	Fields              *[]Field `json:"fields"`
	FollowRequestsCount int64    `json:"follow_requests_count"`
}

// GetAccount return Account.
func (c *Client) GetAccount(ctx context.Context, id string) (*Account, error) {
	var account Account
	params := url.Values{}
	params.Set("with_relationships", strconv.FormatBool(true))
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/accounts/%s", url.PathEscape(string(id))), params, &account, nil)
	if err != nil {
		return nil, err
	}
	if account.Pleroma == nil || len(account.Pleroma.Relationship.ID) < 1 {
		rs, err := c.GetAccountRelationships(ctx, []string{id})
		if err != nil {
			return nil, err
		}
		if len(rs) > 0 {
			account.Pleroma = &AccountPleroma{rs[0]}
		}
	}
	return &account, nil
}

// GetAccountCurrentUser return Account of current user.
func (c *Client) GetAccountCurrentUser(ctx context.Context) (*Account, error) {
	var account Account
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/accounts/verify_credentials", nil, &account, nil)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// Profile is a struct for updating profiles.
type Profile struct {
	// If it is nil it will not be updated.
	// If it is empty, update it with empty.
	DisplayName *string
	Note        *string
	Locked      *bool
	Fields      *[]Field
	Source      *AccountSource

	// Set the base64 encoded character string of the image.
	Avatar *multipart.FileHeader
	Header *multipart.FileHeader
}

// AccountUpdate updates the information of the current user.
func (c *Client) AccountUpdate(ctx context.Context, profile *Profile) (*Account, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if profile.DisplayName != nil {
		err := mw.WriteField("display_name", *profile.DisplayName)
		if err != nil {
			return nil, err
		}
	}
	if profile.Note != nil {
		err := mw.WriteField("note", *profile.Note)
		if err != nil {
			return nil, err
		}
	}
	if profile.Locked != nil {
		err := mw.WriteField("locked", strconv.FormatBool(*profile.Locked))
		if err != nil {
			return nil, err
		}
	}
	if profile.Fields != nil {
		for idx, field := range *profile.Fields {
			err := mw.WriteField(fmt.Sprintf("fields_attributes[%d][name]", idx), field.Name)
			if err != nil {
				return nil, err
			}
			err = mw.WriteField(fmt.Sprintf("fields_attributes[%d][value]", idx), field.Value)
			if err != nil {
				return nil, err
			}
		}
	}
	if profile.Avatar != nil {
		f, err := profile.Avatar.Open()
		if err != nil {
			return nil, err
		}
		fname := filepath.Base(profile.Avatar.Filename)
		part, err := mw.CreateFormFile("avatar", fname)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, f)
		if err != nil {
			return nil, err
		}
	}
	if profile.Header != nil {
		f, err := profile.Header.Open()
		if err != nil {
			return nil, err
		}
		fname := filepath.Base(profile.Header.Filename)
		part, err := mw.CreateFormFile("header", fname)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, f)
		if err != nil {
			return nil, err
		}
	}
	err := mw.Close()
	if err != nil {
		return nil, err
	}
	params := &multipartRequest{Data: &buf, ContentType: mw.FormDataContentType()}
	var account Account
	err = c.doAPI(ctx, http.MethodPatch, "/api/v1/accounts/update_credentials", params, &account, nil)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (c *Client) accountDeleteField(ctx context.Context, field string) (*Account, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_, err := mw.CreateFormField(field)
	if err != nil {
		return nil, err
	}
	err = mw.Close()
	if err != nil {
		return nil, err
	}
	params := &multipartRequest{Data: &buf, ContentType: mw.FormDataContentType()}
	var account Account
	err = c.doAPI(ctx, http.MethodPatch, "/api/v1/accounts/update_credentials", params, &account, nil)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (c *Client) AccountDeleteAvatar(ctx context.Context) (*Account, error) {
	return c.accountDeleteField(ctx, "avatar")
}

func (c *Client) AccountDeleteHeader(ctx context.Context) (*Account, error) {
	return c.accountDeleteField(ctx, "header")
}

// GetAccountStatuses return statuses by specified accuont.
func (c *Client) GetAccountStatuses(ctx context.Context, id string, onlyMedia bool, pg *Pagination) ([]*Status, error) {
	var statuses []*Status
	params := url.Values{}
	params.Set("only_media", strconv.FormatBool(onlyMedia))
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/accounts/%s/statuses", url.PathEscape(string(id))), params, &statuses, pg)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

func (c *Client) getMissingRelationships(ctx context.Context, accounts []*Account) ([]*Account, error) {
	var ids []string
	for _, a := range accounts {
		if a.Pleroma == nil || len(a.Pleroma.Relationship.ID) < 1 {
			ids = append(ids, a.ID)
		}
	}
	if len(ids) < 1 {
		return accounts, nil
	}
	rs, err := c.GetAccountRelationships(ctx, ids)
	if err != nil {
		return nil, err
	}
	rsm := make(map[string]*Relationship, len(rs))
	for _, r := range rs {
		rsm[r.ID] = r
	}
	for _, a := range accounts {
		a.Pleroma = &AccountPleroma{rsm[a.ID]}
	}
	return accounts, nil
}

// GetAccountFollowers return followers list.
func (c *Client) GetAccountFollowers(ctx context.Context, id string, pg *Pagination) ([]*Account, error) {
	var accounts []*Account
	params := url.Values{}
	params.Set("with_relationships", strconv.FormatBool(true))
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/accounts/%s/followers", url.PathEscape(string(id))), params, &accounts, pg)
	if err != nil {
		return nil, err
	}
	accounts, err = c.getMissingRelationships(ctx, accounts)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// GetAccountFollowing return following list.
func (c *Client) GetAccountFollowing(ctx context.Context, id string, pg *Pagination) ([]*Account, error) {
	var accounts []*Account
	params := url.Values{}
	params.Set("with_relationships", strconv.FormatBool(true))
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/accounts/%s/following", url.PathEscape(string(id))), params, &accounts, pg)
	if err != nil {
		return nil, err
	}
	accounts, err = c.getMissingRelationships(ctx, accounts)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// GetBlocks return block list.
func (c *Client) GetBlocks(ctx context.Context, pg *Pagination) ([]*Account, error) {
	var accounts []*Account
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/blocks", nil, &accounts, pg)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// Relationship hold information for relation-ship to the account.
type Relationship struct {
	ID                  string `json:"id"`
	Following           bool   `json:"following"`
	FollowedBy          bool   `json:"followed_by"`
	Blocking            bool   `json:"blocking"`
	BlockedBy           bool   `json:"blocked_by"`
	Muting              bool   `json:"muting"`
	MutingNotifications bool   `json:"muting_notifications"`
	Subscribing         bool   `json:"subscribing"`
	Requested           bool   `json:"requested"`
	DomainBlocking      bool   `json:"domain_blocking"`
	ShowingReblogs      bool   `json:"showing_reblogs"`
	Endorsed            bool   `json:"endorsed"`
}

// AccountFollow follow the account.
func (c *Client) AccountFollow(ctx context.Context, id string, reblogs *bool) (*Relationship, error) {
	var relationship Relationship
	params := url.Values{}
	if reblogs != nil {
		params.Set("reblogs", strconv.FormatBool(*reblogs))
	}
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/accounts/%s/follow", url.PathEscape(id)), params, &relationship, nil)
	if err != nil {
		return nil, err
	}
	return &relationship, nil
}

// AccountUnfollow unfollow the account.
func (c *Client) AccountUnfollow(ctx context.Context, id string) (*Relationship, error) {
	var relationship Relationship
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/accounts/%s/unfollow", url.PathEscape(string(id))), nil, &relationship, nil)
	if err != nil {
		return nil, err
	}
	return &relationship, nil
}

// AccountBlock block the account.
func (c *Client) AccountBlock(ctx context.Context, id string) (*Relationship, error) {
	var relationship Relationship
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/accounts/%s/block", url.PathEscape(string(id))), nil, &relationship, nil)
	if err != nil {
		return nil, err
	}
	return &relationship, nil
}

// AccountUnblock unblock the account.
func (c *Client) AccountUnblock(ctx context.Context, id string) (*Relationship, error) {
	var relationship Relationship
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/accounts/%s/unblock", url.PathEscape(string(id))), nil, &relationship, nil)
	if err != nil {
		return nil, err
	}
	return &relationship, nil
}

// AccountMute mute the account.
func (c *Client) AccountMute(ctx context.Context, id string, notifications bool, duration int) (*Relationship, error) {
	params := url.Values{}
	params.Set("notifications", strconv.FormatBool(notifications))
	params.Set("duration", strconv.Itoa(duration))
	var relationship Relationship
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/accounts/%s/mute", url.PathEscape(string(id))), params, &relationship, nil)
	if err != nil {
		return nil, err
	}
	return &relationship, nil
}

// AccountUnmute unmute the account.
func (c *Client) AccountUnmute(ctx context.Context, id string) (*Relationship, error) {
	var relationship Relationship
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/accounts/%s/unmute", url.PathEscape(string(id))), nil, &relationship, nil)
	if err != nil {
		return nil, err
	}
	return &relationship, nil
}

// GetAccountRelationships return relationship for the account.
func (c *Client) GetAccountRelationships(ctx context.Context, ids []string) ([]*Relationship, error) {
	params := url.Values{}
	for _, id := range ids {
		params.Add("id[]", id)
	}

	var relationships []*Relationship
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/accounts/relationships", params, &relationships, nil)
	if err != nil {
		return nil, err
	}
	return relationships, nil
}

// AccountsSearch search accounts by query.
func (c *Client) AccountsSearch(ctx context.Context, q string, limit int64) ([]*Account, error) {
	params := url.Values{}
	params.Set("q", q)
	params.Set("limit", fmt.Sprint(limit))

	var accounts []*Account
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/accounts/search", params, &accounts, nil)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// FollowRemoteUser send follow-request.
func (c *Client) FollowRemoteUser(ctx context.Context, uri string) (*Account, error) {
	params := url.Values{}
	params.Set("uri", uri)

	var account Account
	err := c.doAPI(ctx, http.MethodPost, "/api/v1/follows", params, &account, nil)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetFollowRequests return follow-requests.
func (c *Client) GetFollowRequests(ctx context.Context, pg *Pagination) ([]*Account, error) {
	var accounts []*Account
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/follow_requests", nil, &accounts, pg)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// FollowRequestAuthorize is authorize the follow request of user with id.
func (c *Client) FollowRequestAuthorize(ctx context.Context, id string) error {
	return c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/follow_requests/%s/authorize", url.PathEscape(string(id))), nil, nil, nil)
}

// FollowRequestReject is rejects the follow request of user with id.
func (c *Client) FollowRequestReject(ctx context.Context, id string) error {
	return c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/follow_requests/%s/reject", url.PathEscape(string(id))), nil, nil, nil)
}

// GetMutes returns the list of users muted by the current user.
func (c *Client) GetMutes(ctx context.Context, pg *Pagination) ([]*Account, error) {
	var accounts []*Account
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/mutes", nil, &accounts, pg)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// Subscribe to receive notifications for all statuses posted by a user
func (c *Client) Subscribe(ctx context.Context, id string) (*Relationship, error) {
	var relationship *Relationship
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/pleroma/accounts/%s/subscribe", url.PathEscape(id)), nil, &relationship, nil)
	if err != nil {
		return nil, err
	}
	return relationship, nil
}

// UnSubscribe to stop receiving notifications from user statuses
func (c *Client) UnSubscribe(ctx context.Context, id string) (*Relationship, error) {
	var relationship *Relationship
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/pleroma/accounts/%s/unsubscribe", url.PathEscape(id)), nil, &relationship, nil)
	if err != nil {
		return nil, err
	}
	return relationship, nil
}

// GetBookmarks returns the list of bookmarked statuses
func (c *Client) GetBookmarks(ctx context.Context, pg *Pagination) ([]*Status, error) {
	var statuses []*Status
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/bookmarks", nil, &statuses, pg)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}
