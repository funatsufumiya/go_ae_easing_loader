package easing_loader

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type EaseType int

const (
    BEZIER EaseType = iota
    LINEAR
    HOLD
)

func easeTypeFromString(s string) EaseType {
    switch s {
    case "bezier":
        return BEZIER
    case "linear":
        return LINEAR
    case "hold":
        return HOLD
    default:
        panic("unknown ease type: " + s)
    }
}

type Ease struct {
    Influence float32
    Speed     float32
}

type Keyframe struct {
    Time            float32
    Value           []float32
    InterpolationOut EaseType
    InterpolationIn  EaseType
    OutEase         Ease
    InEase          Ease
}

type Track struct {
    PropertyName string
    MatchName    string
    ParentName   string
    LayerName    string
    Keyframes    []Keyframe
}

type AEEasingLoader struct {
    Tracks []Track
}

func (ae *AEEasingLoader) LoadJsonFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()
    bytes, err := ioutil.ReadAll(f)
    if err != nil {
        return err
    }
    var raw []map[string]interface{}
    if err := json.Unmarshal(bytes, &raw); err != nil {
        return err
    }
    ae.Tracks = make([]Track, len(raw))
    for i, d := range raw {
        ae.Tracks[i].PropertyName = d["propertyName"].(string)
        ae.Tracks[i].MatchName = d["matchName"].(string)
        if v, ok := d["layerName"]; ok {
            ae.Tracks[i].LayerName = v.(string)
        }
        if v, ok := d["parentName"]; ok {
            ae.Tracks[i].ParentName = v.(string)
        }
        keys := d["keys"].([]interface{})
        for _, k := range keys {
            kmap := k.(map[string]interface{})
            var outEase, inEase Ease
            if oe, ok := kmap["outEase"]; ok && len(oe.([]interface{})) > 0 {
                o := oe.([]interface{})[0].(map[string]interface{})
                if inf, ok := o["influence"].(float64); ok {
                    outEase.Influence = float32(inf)
                }
                if spd, ok := o["speed"].(float64); ok {
                    outEase.Speed = float32(spd)
                }
            }
            if ie, ok := kmap["inEase"]; ok && len(ie.([]interface{})) > 0 {
                i := ie.([]interface{})[0].(map[string]interface{})
                if inf, ok := i["influence"].(float64); ok {
                    inEase.Influence = float32(inf)
                }
                if spd, ok := i["speed"].(float64); ok {
                    inEase.Speed = float32(spd)
                }
            }
            var values []float32
            switch v := kmap["value"].(type) {
            case []interface{}:
                for _, vv := range v {
                    if f, ok := vv.(float64); ok {
                        values = append(values, float32(f))
                    }
                }
            case float64:
                values = []float32{float32(v)}
            }
            var timeVal float32
            if tval, ok := kmap["time"].(float64); ok {
                timeVal = float32(tval)
            }
            ae.Tracks[i].Keyframes = append(ae.Tracks[i].Keyframes, Keyframe{
                Time:            timeVal,
                Value:           values,
                InterpolationOut: easeTypeFromString(kmap["interpolationOut"].(string)),
                InterpolationIn:  easeTypeFromString(kmap["interpolationIn"].(string)),
                OutEase:         outEase,
                InEase:          inEase,
            })
        }
    }
    return nil
}

func (ae *AEEasingLoader) GetPropertyIndex(propertyName, layerName, parentName string) (int, error) {
    for i, t := range ae.Tracks {
        match := (t.PropertyName == propertyName || t.MatchName == propertyName)
        if parentName != "" && layerName != "" {
            if match && t.ParentName == parentName && t.LayerName == layerName {
                return i, nil
            }
        } else if parentName != "" {
            if match && t.ParentName == parentName {
                return i, nil
            }
        } else if layerName != "" {
            if match && t.LayerName == layerName {
                return i, nil
            }
        } else {
            if match {
                return i, nil
            }
        }
    }
    return -1, errors.New("property index not found")
}

func lerp(a, b, t float32) float32 {
    return a + (b-a)*t
}

func cubicBezier(p0, p1, p2, p3, t float32) float32 {
    u := 1.0 - t
    return u*u*u*p0 + 3*u*u*t*p1 + 3*u*t*t*p2 + t*t*t*p3
}

func bezierInterp(
    t, t0, v0 float32, outEase Ease,
    t1, v1 float32, inEase Ease,
) float32 {
    dt := t1 - t0
    if dt <= 0.0 {
        return v0
    }
    localT := (t - t0) / dt
    p0x := float32(0.0)
    p3x := float32(1.0)
    p1x := outEase.Influence / 100.0
    p2x := 1.0 - inEase.Influence / 100.0
    p0y := v0
    p3y := v1
    p1y := v0 + outEase.Speed*dt*(outEase.Influence/100.0)
    p2y := v1 - inEase.Speed*dt*(inEase.Influence/100.0)
    x := localT
    guess := x
    for i := 0; i < 5; i++ {
        bez_x := cubicBezier(p0x, p1x, p2x, p3x, guess)
        bez_dx := 3*(1-guess)*(1-guess)*(p1x-p0x) +
            6*(1-guess)*guess*(p2x-p1x) +
            3*guess*guess*(p3x-p2x)
        if bez_dx == 0.0 {
            break
        }
        guess -= (bez_x - x) / bez_dx
        if guess < 0 {
            guess = 0
        }
        if guess > 1 {
            guess = 1
        }
    }
    t_bez := guess
    return cubicBezier(p0y, p1y, p2y, p3y, t_bez)
}

func (ae *AEEasingLoader) GetValuesAtTime(keys []Keyframe, t float32) []float32 {
    if len(keys) == 0 {
        return nil
    }
    if t <= keys[0].Time {
        return keys[0].Value
    }
    if t >= keys[len(keys)-1].Time {
        return keys[len(keys)-1].Value
    }
    n := len(keys[0].Value)
    result := make([]float32, n)
    var idx1 int
    for i := 1; i < len(keys); i++ {
        if keys[i].Time > t {
            idx1 = i
            break
        }
    }
    idx0 := idx1 - 1
    k0 := keys[idx0]
    k1 := keys[idx1]
    for j := 0; j < n; j++ {
        switch k0.InterpolationOut {
        case HOLD:
            result[j] = k0.Value[j]
        case LINEAR:
            localT := (t - k0.Time) / (k1.Time - k0.Time)
            result[j] = lerp(k0.Value[j], k1.Value[j], localT)
        case BEZIER:
            result[j] = bezierInterp(
                t,
                k0.Time, k0.Value[j], k0.OutEase,
                k1.Time, k1.Value[j], k1.InEase,
            )
        default:
            localT := (t - k0.Time) / (k1.Time - k0.Time)
            result[j] = lerp(k0.Value[j], k1.Value[j], localT)
        }
    }
    return result
}

// Returns the first value (float32) at time t for property index
func (ae *AEEasingLoader) Get(t float32, propertyIndex int) float32 {
    return ae.GetValuesAtTime(ae.Tracks[propertyIndex].Keyframes, t)[0]
}

// Returns the first two values ([x, y]) at time t for property index. If not enough values, missing elements are filled with 0.
func (ae *AEEasingLoader) Get2(t float32, propertyIndex int) (float32, float32) {
    v := ae.GetValuesAtTime(ae.Tracks[propertyIndex].Keyframes, t)
    var out [2]float32
    for i := 0; i < 2 && i < len(v); i++ {
        out[i] = v[i]
    }
    return out[0], out[1]
}

// Returns the first three values ([x, y, z]) at time t for property index. If not enough values, missing elements are filled with 0.
func (ae *AEEasingLoader) Get3(t float32, propertyIndex int) (float32, float32, float32) {
    v := ae.GetValuesAtTime(ae.Tracks[propertyIndex].Keyframes, t)
    var out [3]float32
    for i := 0; i < 3 && i < len(v); i++ {
        out[i] = v[i]
    }
    return out[0], out[1], out[2]
}

// Returns the first four values ([x, y, z, w]) at time t for property index. If not enough values, missing elements are filled with 0.
func (ae *AEEasingLoader) Get4(t float32, propertyIndex int) (float32, float32, float32, float32) {
    v := ae.GetValuesAtTime(ae.Tracks[propertyIndex].Keyframes, t)
    var out [4]float32
    for i := 0; i < 4 && i < len(v); i++ {
        out[i] = v[i]
    }
    return out[0], out[1], out[2], out[3]
}

// Returns all values at time t for property index
func (ae *AEEasingLoader) GetVec(t float32, propertyIndex int) []float32 {
    return ae.GetValuesAtTime(ae.Tracks[propertyIndex].Keyframes, t)
}

// Returns the first value (float32) at time t for property name (with optional layer/parent). If not found, returns 0.
func (ae *AEEasingLoader) GetByName(t float32, propertyName string, layerName string, parentName string) float32 {
    idx, err := ae.GetPropertyIndex(propertyName, layerName, parentName)
    if err != nil {
        return 0
    }
    return ae.Get(t, idx)
}

// Returns the first two values ([x, y]) at time t for property name. If not found, returns (0,0).
func (ae *AEEasingLoader) Get2ByName(t float32, propertyName string, layerName string, parentName string) (float32, float32) {
    idx, err := ae.GetPropertyIndex(propertyName, layerName, parentName)
    if err != nil {
        return 0, 0
    }
    return ae.Get2(t, idx)
}

// Returns the first three values ([x, y, z]) at time t for property name. If not found, returns (0,0,0).
func (ae *AEEasingLoader) Get3ByName(t float32, propertyName string, layerName string, parentName string) (float32, float32, float32) {
    idx, err := ae.GetPropertyIndex(propertyName, layerName, parentName)
    if err != nil {
        return 0, 0, 0
    }
    return ae.Get3(t, idx)
}

// Returns the first four values ([x, y, z, w]) at time t for property name. If not found, returns (0,0,0,0).
func (ae *AEEasingLoader) Get4ByName(t float32, propertyName string, layerName string, parentName string) (float32, float32, float32, float32) {
    idx, err := ae.GetPropertyIndex(propertyName, layerName, parentName)
    if err != nil {
        return 0, 0, 0, 0
    }
    return ae.Get4(t, idx)
}

// Returns all values at time t for property name. If not found, returns nil.
func (ae *AEEasingLoader) GetVecByName(t float32, propertyName string, layerName string, parentName string) []float32 {
    idx, err := ae.GetPropertyIndex(propertyName, layerName, parentName)
    if err != nil {
        return nil
    }
    return ae.GetVec(t, idx)
}