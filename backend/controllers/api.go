package controllers

import (
	"encoding/json"
	"fmt"
	"linkdock/models"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type DataController struct {
	web.Controller
}

// RawSeedCategory 用于从 JSON 文件读取初始数据模板
type RawSeedCategory struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	IconName string `json:"iconName"`
}

type RawSeedSite struct {
	Id          string `json:"id"`
	CategoryId  string `json:"categoryId"`
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sortOrder"`
}

type SeedConfig struct {
	Categories []RawSeedCategory `json:"categories"`
	Sites      []RawSeedSite     `json:"sites"`
}

type SiteOrderItem struct {
	Id         string `json:"id"`
	CategoryId string `json:"categoryId"`
	SortOrder  int    `json:"sortOrder"`
}

type siteStateRequest struct {
	IsFavorite     *bool  `json:"isFavorite"`
	WorkflowStatus string `json:"workflowStatus"`
}

func readSeedDataFromFile() ([]RawSeedCategory, []RawSeedSite) {
	data, err := readSeedFile()
	if err != nil {
		return nil, nil
	}

	var seed SeedConfig
	if err := json.Unmarshal(data, &seed); err != nil {
		return nil, nil
	}

	return seed.Categories, seed.Sites
}

func seedFilePaths() []string {
	return []string{"conf/seed.json", "backend/conf/seed.json", "/app/backend/conf/seed.json"}
}

func resolveSeedFilePath() string {
	for _, p := range seedFilePaths() {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "conf/seed.json"
}

func readSeedFile() ([]byte, error) {
	var lastErr error
	for _, p := range seedFilePaths() {
		data, err := os.ReadFile(p)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func validateSeedConfig(seed SeedConfig) error {
	if len(seed.Categories) == 0 {
		return fmt.Errorf("至少需要保留一个默认分类")
	}

	categoryIds := make(map[string]bool)
	for _, category := range seed.Categories {
		categoryId := strings.TrimSpace(category.Id)
		if categoryId == "" {
			return fmt.Errorf("分类 ID 不能为空")
		}
		if categoryIds[categoryId] {
			return fmt.Errorf("分类 ID 不能重复：%s", categoryId)
		}
		if strings.TrimSpace(category.Name) == "" {
			return fmt.Errorf("分类名称不能为空")
		}
		categoryIds[categoryId] = true
	}

	siteIds := make(map[string]bool)
	for _, site := range seed.Sites {
		siteId := strings.TrimSpace(site.Id)
		if siteId == "" {
			return fmt.Errorf("书签 ID 不能为空")
		}
		if siteIds[siteId] {
			return fmt.Errorf("书签 ID 不能重复：%s", siteId)
		}
		if strings.TrimSpace(site.Name) == "" {
			return fmt.Errorf("书签名称不能为空")
		}
		if _, err := normalizeSiteURL(site.Url); err != nil {
			return fmt.Errorf("书签网址无效：%s", site.Name)
		}
		if !categoryIds[strings.TrimSpace(site.CategoryId)] {
			return fmt.Errorf("书签「%s」引用了不存在的分类", site.Name)
		}
		siteIds[siteId] = true
	}

	return nil
}

func (c *DataController) GetSeedData() {
	if !IsSuperAdminHelper(&c.Controller) {
		c.Ctx.Output.SetStatus(http.StatusForbidden)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "只有超级管理员可以维护默认数据"}
		c.ServeJSON()
		return
	}

	data, err := readSeedFile()
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "读取默认数据失败"}
		c.ServeJSON()
		return
	}

	var seed SeedConfig
	if err := json.Unmarshal(data, &seed); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "默认数据格式错误"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"success":    true,
		"categories": seed.Categories,
		"sites":      seed.Sites,
	}
	c.ServeJSON()
}

func (c *DataController) SaveSeedData() {
	if !IsSuperAdminHelper(&c.Controller) {
		c.Ctx.Output.SetStatus(http.StatusForbidden)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "只有超级管理员可以维护默认数据"}
		c.ServeJSON()
		return
	}

	var seed SeedConfig
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &seed); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请求格式错误"}
		c.ServeJSON()
		return
	}

	for i := range seed.Categories {
		seed.Categories[i].Id = strings.TrimSpace(seed.Categories[i].Id)
		seed.Categories[i].Name = strings.TrimSpace(seed.Categories[i].Name)
		seed.Categories[i].IconName = strings.TrimSpace(seed.Categories[i].IconName)
		if seed.Categories[i].IconName == "" {
			seed.Categories[i].IconName = "📁"
		}
	}
	for i := range seed.Sites {
		seed.Sites[i].Id = strings.TrimSpace(seed.Sites[i].Id)
		seed.Sites[i].CategoryId = strings.TrimSpace(seed.Sites[i].CategoryId)
		seed.Sites[i].Name = strings.TrimSpace(seed.Sites[i].Name)
		seed.Sites[i].Url = strings.TrimSpace(seed.Sites[i].Url)
		seed.Sites[i].Description = strings.TrimSpace(seed.Sites[i].Description)
		seed.Sites[i].Icon = strings.TrimSpace(seed.Sites[i].Icon)
		seed.Sites[i].SortOrder = i
	}

	if err := validateSeedConfig(seed); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": err.Error()}
		c.ServeJSON()
		return
	}

	nextData, err := json.MarshalIndent(seed, "", "  ")
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "序列化默认数据失败"}
		c.ServeJSON()
		return
	}

	seedPath := resolveSeedFilePath()
	if err := os.WriteFile(seedPath, append(nextData, '\n'), 0644); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "保存默认数据失败"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{"success": true, "message": "默认数据已保存"}
	c.ServeJSON()
}

func (c *DataController) GetPublicData() {
	rawCats, rawSites := readSeedDataFromFile()

	var cats []*models.Category
	var sites []*models.Site

	// 为公共数据映射临时数字 ID
	for i, rc := range rawCats {
		cats = append(cats, &models.Category{
			Id:       int64(i + 1),
			Name:     rc.Name,
			IconName: rc.IconName,
		})
	}
	for i, rs := range rawSites {
		catIdx := 0
		for idx, rc := range rawCats {
			if rc.Id == rs.CategoryId {
				catIdx = idx + 1
				break
			}
		}
		sortOrder := rs.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		sites = append(sites, &models.Site{
			Id:          int64(i + 1),
			CategoryId:  int64(catIdx),
			SortOrder:   sortOrder,
			Name:        rs.Name,
			Url:         rs.Url,
			Description: rs.Description,
			LogoUrl:     rs.Icon,
		})
	}

	c.Data["json"] = map[string]interface{}{
		"categories": cats,
		"sites":      sites,
	}
	c.ServeJSON()
}

// seedDefaultData 为新注册用户将模板数据持久化到数据库
func seedDefaultData(userId int64) {
	o := orm.NewOrm()
	rawCats, rawSites := readSeedDataFromFile()

	// 映射表：seed_string_id -> db_actual_id
	catIdMap := make(map[string]int64)
	defaultSeedCategory := RawSeedCategory{
		Id:       "default",
		Name:     "默认分类",
		IconName: "📁",
	}
	for _, rc := range rawCats {
		if rc.Id == "default" {
			defaultSeedCategory = rc
			break
		}
	}

	// 1. 创建默认分类
	defaultCat := &models.Category{
		UserId:    userId,
		Name:      defaultSeedCategory.Name,
		IconName:  defaultSeedCategory.IconName,
		SortOrder: 0,
	}
	if id, err := o.Insert(defaultCat); err == nil {
		catIdMap["default"] = id
	}

	// 2. 注入种子分类
	nextCategorySortOrder := 1
	for _, rc := range rawCats {
		if rc.Id == "default" {
			continue
		}
		newCat := models.Category{
			UserId:    userId,
			Name:      rc.Name,
			IconName:  rc.IconName,
			SortOrder: nextCategorySortOrder,
		}
		if id, err := o.Insert(&newCat); err == nil {
			catIdMap[rc.Id] = id
		}
		nextCategorySortOrder++
	}

	// 3. 注入种子书签
	for i, rs := range rawSites {
		sortOrder := rs.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		newSite := models.Site{
			UserId:         userId,
			CategoryId:     catIdMap[rs.CategoryId],
			SortOrder:      sortOrder,
			Name:           rs.Name,
			Url:            rs.Url,
			Description:    rs.Description,
			LogoUrl:        rs.Icon,
			WorkflowStatus: "none",
		}
		o.Insert(&newSite)
	}
}

// UpdateCategoriesOrder 更新分类排序
func (c *DataController) UpdateCategoriesOrder() {
	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请先登录"}
		c.ServeJSON()
		return
	}

	var data struct {
		Ids []string `json:"ids"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &data); err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "参数错误"}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	for i, idStr := range data.Ids {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		// 确保只更新属于该用户的分类
		o.QueryTable(new(models.Category)).Filter("id", id).Filter("user_id", userId).Update(orm.Params{
			"sort_order": i,
		})
	}

	c.Data["json"] = map[string]interface{}{"success": true}
	c.ServeJSON()
}

// GetUserData 接口
func (c *DataController) GetUserData() {
	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请先登录"}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	var categories []models.Category
	var sites []models.Site

	// 使用明确的字段名进行过滤，并按照 sort_order 排序
	o.QueryTable(new(models.Category)).Filter("user_id", userId).OrderBy("sort_order", "id").All(&categories)

	if len(categories) == 0 {
		seedDefaultData(userId)
		// 重新拉取一次
		o.QueryTable(new(models.Category)).Filter("user_id", userId).OrderBy("sort_order", "id").All(&categories)
	}

	o.QueryTable(new(models.Site)).Filter("user_id", userId).OrderBy("category_id", "sort_order", "id").All(&sites)

	c.Data["json"] = map[string]interface{}{
		"categories": categories,
		"sites":      sites,
	}
	c.ServeJSON()
}

func (c *DataController) CreateSite() {
	var site models.Site
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &site); err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "数据解析失败"}
		c.ServeJSON()
		return
	}

	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "未登录"}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	normalizedURL, err := normalizeSiteURL(site.Url)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请输入有效的网址"}
		c.ServeJSON()
		return
	}

	hasDuplicate, err := userHasDuplicateSiteURL(o, userId, normalizedURL, 0)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "检查书签重复时失败"}
		c.ServeJSON()
		return
	}
	if hasDuplicate {
		c.Ctx.Output.SetStatus(http.StatusConflict)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "这个网址已经收藏过了"}
		c.ServeJSON()
		return
	}

	site.UserId = userId
	site.Url = normalizedURL
	site.TagsText = strings.TrimSpace(site.TagsText)
	site.WorkflowStatus = sanitizeWorkflowStatus(site.WorkflowStatus, true)
	site.SortOrder = getNextSiteSortOrder(o, userId, site.CategoryId)
	if _, err := o.Insert(&site); err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": err.Error()}
	} else {
		c.Data["json"] = map[string]interface{}{"success": true}
	}
	c.ServeJSON()
}

func (c *DataController) CreateCategory() {
	var cat models.Category
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &cat); err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "数据解析失败"}
		c.ServeJSON()
		return
	}

	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "未登录"}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	cat.UserId = userId
	if _, err := o.Insert(&cat); err != nil {
		c.Data["json"] = map[string]bool{"success": false}
	} else {
		c.Data["json"] = map[string]bool{"success": true}
	}
	c.ServeJSON()
}

func getNextSiteSortOrder(o orm.Ormer, userId int64, categoryId int64) int {
	var lastSite models.Site
	err := o.QueryTable(new(models.Site)).
		Filter("user_id", userId).
		Filter("category_id", categoryId).
		OrderBy("-sort_order", "-id").
		One(&lastSite)
	if err == orm.ErrNoRows {
		return 0
	}
	if err != nil {
		return 0
	}
	return lastSite.SortOrder + 1
}

func sanitizeWorkflowStatus(status string, defaultToUnorganized bool) string {
	switch strings.TrimSpace(status) {
	case "favorite":
		return "favorite"
	case "read_later":
		return "read_later"
	case "unorganized":
		return "unorganized"
	case "none":
		return "none"
	default:
		if defaultToUnorganized {
			return "unorganized"
		}
		return "none"
	}
}

func normalizeSiteURL(raw string) (string, error) {
	target := strings.TrimSpace(raw)
	if target == "" {
		return "", fmt.Errorf("empty url")
	}
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}

	parsed, err := url.Parse(target)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("invalid url")
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	if parsed.Path == "/" {
		parsed.Path = ""
	}

	return parsed.String(), nil
}

func userHasDuplicateSiteURL(o orm.Ormer, userId int64, normalizedURL string, excludeSiteID int64) (bool, error) {
	var existingSites []models.Site
	_, err := o.QueryTable(new(models.Site)).
		Filter("user_id", userId).
		All(&existingSites, "id", "url")
	if err != nil {
		return false, err
	}

	for _, existing := range existingSites {
		if excludeSiteID != 0 && existing.Id == excludeSiteID {
			continue
		}

		existingURL, err := normalizeSiteURL(existing.Url)
		if err != nil {
			continue
		}
		if existingURL == normalizedURL {
			return true, nil
		}
	}

	return false, nil
}

type FetchMetaRequest struct {
	Url string `json:"url"`
}

func (c *DataController) FetchMetadata() {
	var req FetchMetaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请求格式错误"}
		c.ServeJSON()
		return
	}

	targetUrl := strings.TrimSpace(req.Url)
	if !strings.HasPrefix(targetUrl, "http://") && !strings.HasPrefix(targetUrl, "https://") {
		targetUrl = "http://" + targetUrl
	}

	parsedUrl, err := url.Parse(targetUrl)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "无效的 URL"}
		c.ServeJSON()
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	httpRequest, _ := http.NewRequest("GET", targetUrl, nil)
	httpRequest.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	res, err := client.Do(httpRequest)
	if err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "无法访问该网站"}
		c.ServeJSON()
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "解析网页失败"}
		c.ServeJSON()
		return
	}

	title := strings.TrimSpace(doc.Find("title").First().Text())
	desc, _ := doc.Find("meta[name='description']").Attr("content")

	iconUrl, _ := doc.Find("link[rel='icon'], link[rel='shortcut icon']").Attr("href")
	if iconUrl != "" {
		if !strings.HasPrefix(iconUrl, "http") {
			if strings.HasPrefix(iconUrl, "//") {
				iconUrl = parsedUrl.Scheme + ":" + iconUrl
			} else if strings.HasPrefix(iconUrl, "/") {
				iconUrl = parsedUrl.Scheme + "://" + parsedUrl.Host + iconUrl
			} else {
				iconUrl = parsedUrl.Scheme + "://" + parsedUrl.Host + "/" + iconUrl
			}
		}
	} else {
		iconUrl = "https://www.google.com/s2/favicons?domain=" + parsedUrl.Host + "&sz=128"
	}

	c.Data["json"] = map[string]interface{}{
		"success":     true,
		"title":       title,
		"description": strings.TrimSpace(desc),
		"icon":        iconUrl,
	}
	c.ServeJSON()
}

func (c *DataController) UpdateSite() {
	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "未登录"}
		c.ServeJSON()
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var req models.Site
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err == nil {
		o := orm.NewOrm()
		site := models.Site{Id: id, UserId: userId}
		if o.Read(&site, "Id", "UserId") == nil {
			normalizedURL, err := normalizeSiteURL(req.Url)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = map[string]interface{}{"success": false, "message": "请输入有效的网址"}
				c.ServeJSON()
				return
			}

			hasDuplicate, err := userHasDuplicateSiteURL(o, userId, normalizedURL, site.Id)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusInternalServerError)
				c.Data["json"] = map[string]interface{}{"success": false, "message": "检查书签重复时失败"}
				c.ServeJSON()
				return
			}
			if hasDuplicate {
				c.Ctx.Output.SetStatus(http.StatusConflict)
				c.Data["json"] = map[string]interface{}{"success": false, "message": "这个网址已经收藏过了"}
				c.ServeJSON()
				return
			}

			categoryChanged := site.CategoryId != req.CategoryId
			site.Name = req.Name
			site.Url = normalizedURL
			site.Description = req.Description
			site.CategoryId = req.CategoryId
			site.LogoUrl = req.LogoUrl
			site.TagsText = strings.TrimSpace(req.TagsText)
			site.IsFavorite = req.IsFavorite
			site.WorkflowStatus = sanitizeWorkflowStatus(req.WorkflowStatus, false)
			if categoryChanged {
				site.SortOrder = getNextSiteSortOrder(o, userId, req.CategoryId)
			}

			if _, err := o.Update(&site); err != nil {
				c.Data["json"] = map[string]interface{}{"success": false, "message": "更新失败"}
			} else {
				c.Data["json"] = map[string]interface{}{"success": true}
			}
		} else {
			c.Data["json"] = map[string]interface{}{"success": false, "message": "无权操作该数据"}
		}
	} else {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请求内容解析失败"}
	}
	c.ServeJSON()
}

func (c *DataController) UpdateSitesOrder() {
	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请先登录"}
		c.ServeJSON()
		return
	}

	var data struct {
		Sites []SiteOrderItem `json:"sites"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &data); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "参数错误"}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	for _, item := range data.Sites {
		siteId, err := strconv.ParseInt(item.Id, 10, 64)
		if err != nil {
			continue
		}
		categoryId, err := strconv.ParseInt(item.CategoryId, 10, 64)
		if err != nil {
			continue
		}

		_, err = o.QueryTable(new(models.Site)).
			Filter("id", siteId).
			Filter("user_id", userId).
			Update(orm.Params{
				"category_id": categoryId,
				"sort_order":  item.SortOrder,
			})
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = map[string]interface{}{"success": false, "message": "更新书签排序失败"}
			c.ServeJSON()
			return
		}
	}

	c.Data["json"] = map[string]interface{}{"success": true}
	c.ServeJSON()
}

func (c *DataController) TrackSiteVisit() {
	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请先登录"}
		c.ServeJSON()
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	o := orm.NewOrm()
	site := models.Site{Id: id, UserId: userId}
	if err := o.Read(&site, "Id", "UserId"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "书签不存在"}
		c.ServeJSON()
		return
	}

	now := time.Now()
	site.VisitCount++
	site.LastVisitedAt = &now
	if _, err := o.Update(&site, "VisitCount", "LastVisitedAt"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "更新访问记录失败"}
	} else {
		c.Data["json"] = map[string]interface{}{"success": true, "visitCount": site.VisitCount, "lastVisitedAt": now}
	}
	c.ServeJSON()
}

func (c *DataController) UpdateSiteState() {
	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请先登录"}
		c.ServeJSON()
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var req siteStateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请求格式错误"}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	site := models.Site{Id: id, UserId: userId}
	if err := o.Read(&site, "Id", "UserId"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "书签不存在"}
		c.ServeJSON()
		return
	}

	fields := make([]string, 0, 2)
	if req.IsFavorite != nil {
		site.IsFavorite = *req.IsFavorite
		fields = append(fields, "IsFavorite")
	}
	if req.WorkflowStatus != "" {
		site.WorkflowStatus = sanitizeWorkflowStatus(req.WorkflowStatus, false)
		fields = append(fields, "WorkflowStatus")
	}

	if len(fields) == 0 {
		c.Data["json"] = map[string]interface{}{"success": true}
		c.ServeJSON()
		return
	}

	if _, err := o.Update(&site, fields...); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "更新书签状态失败"}
	} else {
		c.Data["json"] = map[string]interface{}{
			"success":        true,
			"isFavorite":     site.IsFavorite,
			"workflowStatus": site.WorkflowStatus,
		}
	}
	c.ServeJSON()
}

func (c *DataController) DeleteSite() {
	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "未登录"}
		c.ServeJSON()
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	o := orm.NewOrm()
	if _, err := o.QueryTable(new(models.Site)).Filter("Id", id).Filter("UserId", userId).Delete(); err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "删除失败"}
	} else {
		c.Data["json"] = map[string]bool{"success": true}
	}
	c.ServeJSON()
}

func (c *DataController) UpdateCategory() {
	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "未登录"}
		c.ServeJSON()
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var req models.Category
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err == nil {
		o := orm.NewOrm()
		cat := models.Category{Id: id, UserId: userId}
		if o.Read(&cat, "Id", "UserId") == nil {
			cat.Name = req.Name
			cat.IconName = req.IconName
			o.Update(&cat)
			c.Data["json"] = map[string]bool{"success": true}
		} else {
			c.Data["json"] = map[string]interface{}{"success": false, "message": "无权操作该数据"}
		}
	} else {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请求内容解析失败"}
	}
	c.ServeJSON()
}

func (c *DataController) DeleteCategory() {
	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "未登录"}
		c.ServeJSON()
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	o := orm.NewOrm()
	if _, err := o.QueryTable(new(models.Category)).Filter("Id", id).Filter("UserId", userId).Delete(); err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "删除失败"}
	} else {
		o.QueryTable(new(models.Site)).Filter("CategoryId", id).Filter("UserId", userId).Delete()
		c.Data["json"] = map[string]bool{"success": true}
	}
	c.ServeJSON()
}
