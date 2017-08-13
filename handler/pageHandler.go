package handler

import (
	"fmt"
	"math"

	"../model"
)

//分页方法，根据传递过来的页数，总数，返回分页的内容 7个页数 前 1，2，3，4，5 后 的格式返回,小于5页返回具体页数
/**
@param page 页码
@param nums 总数
**/
func Paginator(page int, nums int64) map[string]interface{} {

	paginatorMap := make(map[string]interface{})
	if nums == 0 {
		paginatorMap["pages"] = 0
		paginatorMap["totals"] = 0
		paginatorMap["totalpages"] = 0
		paginatorMap["firstpage"] = 0
		paginatorMap["lastpage"] = 0
		paginatorMap["currpage"] = 0
		paginatorMap["startIndex"] = 0
	}

	var firstpage int //前一页地址
	var lastpage int  //后一页地址
	prepage := model.PAGE_SIZE
	//根据nums总数，和prepage每页数量 生成分页总数
	totalpages := int(math.Ceil(float64(nums) / float64(prepage))) //page总数
	if page > totalpages {
		page = totalpages
	}
	if page <= 0 {
		page = 1
	}
	var pages []int
	switch {
	case page >= totalpages-5 && totalpages > 5: //最后5页
		start := totalpages - 5 + 1
		firstpage = page - 1
		lastpage = int(math.Min(float64(totalpages), float64(page+1)))
		pages = make([]int, 5)
		for i := range pages {
			pages[i] = start + i
		}
	case page >= 3 && totalpages > 5:
		start := page - 3 + 1
		pages = make([]int, 5)
		firstpage = page - 3
		for i := range pages {
			pages[i] = start + i
		}
		firstpage = page - 1
		lastpage = page + 1
	default:
		pages = make([]int, int(math.Min(5, float64(totalpages))))
		for i := range pages {
			pages[i] = i + 1
		}
		firstpage = page - 1
		if firstpage < 1 {
			firstpage = 0
		}
		lastpage = page + 1
		if lastpage > totalpages {
			lastpage = totalpages
		}
	}
	//数据库分页数据
	startIndex := ((page - 1) * model.PAGE_SIZE)
	paginatorMap["pages"] = pages
	paginatorMap["totals"] = nums
	paginatorMap["totalpages"] = totalpages
	paginatorMap["firstpage"] = firstpage
	paginatorMap["lastpage"] = lastpage
	paginatorMap["currpage"] = page
	paginatorMap["startIndex"] = startIndex
	fmt.Printf("分叶数据：%s", paginatorMap)
	return paginatorMap
}
