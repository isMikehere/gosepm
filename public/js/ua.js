/**
 * Created by mike on 2017/9/4.
 */
//init the ranking chart
//******************************************************************************
var echart1 = echarts.init(document.getElementById('chart1'));
echart1.showLoading({
    text: '正在努力加载中...'
});

var values = [];
// 同步执行
$.ajaxSettings.async = false;

// 加载数据
var uid = $("#uid").val();
var nickName = $("#nickName").val();
var path = $("#webpath").val() + '/user/trxRate/' + uid
$.getJSON(path, function (json) {
    if (json.code == "200") {
        values = json.data;
        echart1.hideLoading()
    }
});
// 指定图表的配置项和数据
option = {
    title: {
        text: nickName + '的交易情况',
        subtext: '',
        x: 'center'
    },
    tooltip: {
        trigger: 'item',
        formatter: "{a} <br/>{b} : {c} ({d}%)"
    },
    legend: {
        orient: 'vertical',
        left: 'left',
        data: ['盈利笔数', '亏损笔数']
    },
    series: [
        {
            name: nickName + '的交易情况',
            type: 'pie',
            radius: '30%',
            center: ['30%', '35%'],
            data: values,
            itemStyle: {
                emphasis: {
                    shadowBlur: 10,
                    shadowOffsetX: 0,
                    shadowColor: 'rgba(0, 0, 0, 0.5)'
                }
            }
        }
    ]
};
// 使用刚指定的配置项和数据显示图表。
echart1.setOption(option);
//******************************************************************************
var echart2 = echarts.init(document.getElementById('chart2'));
$.getJSON($("#webpath").val() + '/user/rankDataChart/' + uid, function (json) {

    var data = json.data

    var myRegression = ecStat.regression('exponential', data);
    myRegression.points.sort(function (a, b) {
        return a[0] - b[0];
    });

    echart2.setOption({
        title: {
            text: nickName + '周排名变化',
            left: 'center',
            top: 21
        },
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'cross'
            }
        },
        xAxis: {
            type: 'value',
            min: 10,
            splitLine: {
                lineStyle: {
                    type: 'dashed'
                }
            },
        },
        yAxis: {
            type: 'value',
            min: 1,
            splitLine: {
                lineStyle: {
                    type: 'dashed'
                }
            },
        },
        series: [{
            name: '周',
            type: 'scatter',
            show: false,
            label: {
                emphasis: {
                    show: false
                }
            },
            data: data
        }, {
            name: '排名',
            type: 'line',
            showSymbol: false,
            data: myRegression.points,
            markPoint: {
                itemStyle: {
                    normal: {
                        color: 'transparent'
                    }
                },
                label: {
                    normal: {
                        show: false,
                        position: 'left',
                        formatter: myRegression.expression,
                        textStyle: {
                            color: '#333',
                            fontSize: 14
                        }
                    }
                },
                data: [{
                    coord: myRegression.points[myRegression.points.length - 1]
                }]
            }
        }]
    });
});