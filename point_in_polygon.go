package main

import (
	"fmt"
)

// Point GCJ-02 坐标点
type Point struct {
	Lng float64 // 经度
	Lat float64 // 纬度
}

// IsPointInPolygon 判断点是否在多边形内（射线法）
// polygon: 多边形顶点（有序，顺时针或逆时针均可）
func IsPointInPolygon(point Point, polygon []Point) bool {
	n := len(polygon)
	if n < 3 {
		return false
	}

	inside := false
	j := n - 1 // 最后一个顶点

	for i := 0; i < n; i++ {
		// 射线：从 point 向右水平延伸
		// 检查边 (polygon[i], polygon[j]) 是否与射线相交
		if ((polygon[i].Lat > point.Lat) != (polygon[j].Lat > point.Lat)) &&
			(point.Lng < (polygon[j].Lng-polygon[i].Lng)*(point.Lat-polygon[i].Lat)/(polygon[j].Lat-polygon[i].Lat)+polygon[i].Lng) {
			inside = !inside // 相交次数翻转
		}
		j = i
	}

	return inside
}

// IsPointOnEdge 判断点是否在多边形边上
func IsPointOnEdge(point Point, polygon []Point, epsilon float64) bool {
	n := len(polygon)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		if isPointOnLineSegment(point, polygon[i], polygon[j], epsilon) {
			return true
		}
	}
	return false
}

// isPointOnLineSegment 判断点是否在线段上
func isPointOnLineSegment(point, lineStart, lineEnd Point, epsilon float64) bool {
	// 检查点是否在线段的包围盒内
	if point.Lng < min(lineStart.Lng, lineEnd.Lng)-epsilon ||
		point.Lng > max(lineStart.Lng, lineEnd.Lng)+epsilon ||
		point.Lat < min(lineStart.Lat, lineEnd.Lat)-epsilon ||
		point.Lat > max(lineStart.Lat, lineEnd.Lat)+epsilon {
		return false
	}

	// 计算点到直线的距离
	cross := (point.Lat-lineStart.Lat)*(lineEnd.Lng-lineStart.Lng) - (point.Lng-lineStart.Lng)*(lineEnd.Lat-lineStart.Lat)
	if abs(cross) > epsilon {
		return false
	}

	return true
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	// 示例：不规则多边形（五边形）
	polygon := []Point{
		{116.397128, 39.916527}, // 天安门
		{116.417128, 39.916527}, // 东
		{116.417128, 39.936527}, // 北
		{116.397128, 39.936527}, // 西北
		{116.387128, 39.926527}, // 西南
	}

	// 测试点
	testPoints := []struct {
		point Point
		desc  string
	}{
		{Point{116.407128, 39.926527}, "中心点（应该在内部）"},
		{Point{116.397128, 39.916527}, "顶点（在边上）"},
		{Point{116.380000, 39.920000}, "外部点（西侧）"},
		{Point{116.420000, 39.940000}, "外部点（东北）"},
	}

	fmt.Println("=== 点在不规则多边形内判断 ===\n")

	for _, tp := range testPoints {
		inside := IsPointInPolygon(tp.point, polygon)
		onEdge := IsPointOnEdge(tp.point, polygon, 1e-9)

		status := "❌ 外部"
		if onEdge {
			status = "📍 边上"
		} else if inside {
			status = "✅ 内部"
		}

		fmt.Printf("%s\n", tp.desc)
		fmt.Printf("  坐标: (%.6f, %.6f)\n", tp.point.Lng, tp.point.Lat)
		fmt.Printf("  结果: %s\n\n", status)
	}
}
