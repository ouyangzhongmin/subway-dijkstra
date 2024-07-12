package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/RyanCarrier/dijkstra/v2"
)

// 地铁数据
type Subway struct {
	S string       `json:"s"` //名称
	I string       `json:"i"` //id
	L []SubwayLine `json:"l"` // 线路数据
}

// 地铁线路
type SubwayLine struct {
	KN string           `json:"kn"`
	LN string           `json:"ln"`
	ST []SubwayStations `json:"st"`
}

// 地铁站点
type SubwayStations struct {
	Rs    string `json:"rs"`
	Udpx  string `json:"udpx"`
	Su    string `json:"su"`
	En    string `json:"en"`
	N     string `json:"n"`
	Sid   string `json:"sid"`
	P     string `json:"p"`
	R     string `json:"r"`
	Udsi  string `json:"udsi"`
	T     string `json:"t"`
	Si    string `json:"si"`
	Sl    string `json:"sl"`
	Udli  string `json:"udli"`
	Poiid string `json:"poiid"`
	Lg    string `json:"lg"`
	Sp    string `json:"sp"`
}

type subwaydata struct {
	Line string  `json:"line"`
	Name string  `json:"name"`
	Lo   float64 `json:"lo"` //经度
	La   float64 `json:"la"` //纬度
}

func main() {
	// 在线数据 基于高德地图平台获取
	//resp, err := http.Get("http://map.amap.com/service/subway?_1608683463948&srhdata=4201_drw_wuhan.json")
	//if err != nil {
	//	panic(err)
	//}
	//defer resp.Body.Close()
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	panic(err)
	//}
	//测试直接读本地文件
	body, err := ReadFile("./wuhan-subway-json.txt")
	if err != nil {
		panic(err)
	}
	var subwayData Subway
	err = json.Unmarshal(body, &subwayData)
	if err != nil {
		panic(err)
	}

	//全部的线路数据
	totalLines := make([]subwaydata, 0)
	// 索引线路映射
	verticesMap := make(map[string]int)
	//待构建的索引线路数据
	vertices := make([]subwaydata, 0)
	for i := 0; i < len(subwayData.L); i++ {
		for j := 0; j < len(subwayData.L[i].ST); j++ {
			var s subwaydata
			s.Line = subwayData.L[i].LN
			s.Name = subwayData.L[i].ST[j].N
			sl := strings.Split(subwayData.L[i].ST[j].Sl, ",")
			lo, err := strconv.ParseFloat(sl[0], 64)
			if err != nil {
				panic(err)
			}
			la, err := strconv.ParseFloat(sl[1], 64)
			if err != nil {
				panic(err)
			}
			s.Lo = lo
			s.La = la
			totalLines = append(totalLines, s)
			//过滤重复的站点数据
			if _, ok := verticesMap[s.Name]; !ok {
				vertices = append(vertices, s)
				verticesMap[s.Name] = len(vertices) - 1
			}
		}
	}
	fmt.Println("地铁数据:", totalLines)
	//构建图关联映射, 这个库使用的事邻接矩阵的方式实现的带权图
	arcs := make(map[int]map[int]uint64)
	for i := 0; i < len(totalLines); i++ {
		j := i + 1
		if j >= len(totalLines) {
			continue
		}
		var dist float64 = 0
		if totalLines[i].Line == totalLines[j].Line {
			dist = Distance(totalLines[i].Lo, totalLines[i].La, totalLines[j].Lo, totalLines[j].La)
			vi := verticesMap[totalLines[i].Name]
			vj := verticesMap[totalLines[j].Name]
			if _, ok := arcs[vi]; !ok {
				m := make(map[int]uint64)
				m[vj] = uint64(dist)
				arcs[vi] = m
			} else {
				arcs[vi][vj] = uint64(dist)
			}
		}
	}
	graph := initGraph(vertices, arcs)

	dest := verticesMap["十里铺"]
	src := verticesMap["汉口火车站"]
	best, err := graph.ShortestAll(src, dest)
	if err != nil {
		fmt.Println(graph)
		log.Fatal(err)
	}
	fmt.Println("Shortest distances are", best.Distance, "with paths; ", best.Paths)
	for i := 0; i < len(best.Paths); i++ {
		fmt.Println(i, "路线::")
		for j := 0; j < len(best.Paths[i]); j++ {
			fmt.Println(vertices[best.Paths[i][j]].Name)
		}
	}
}

func initGraph(vertices []subwaydata, arcs map[int]map[int]uint64) dijkstra.Graph {
	graph := dijkstra.NewGraph()
	for i := 0; i < len(vertices); i++ {
		//Add the  verticies
		arc := arcs[i]
		graph.AddVertexAndArcs(i, arc)
	}
	return graph
}

// 地球半径（米）
const EarthRadius = 6371000

// 计算两点之间的距离
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	radLat1 := lat1 * math.Pi / 180.0
	radLat2 := lat2 * math.Pi / 180.0
	a := radLat1 - radLat2
	b := (lon1 - lon2) * math.Pi / 180.0

	alpha := math.Sin(a/2)*math.Sin(a/2) +
		math.Cos(radLat1)*math.Cos(radLat2)*
			math.Sin(b/2)*math.Sin(b/2)
	c := 2 * math.Atan2(math.Sqrt(alpha), math.Sqrt(1-alpha))
	return EarthRadius * c
}

func ReadFile(filePath string) ([]byte, error) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}
