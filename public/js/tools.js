/**
 * Created by mike on 2017/9/3.
 */

/**
 *自定义除法
 * @param val
 * @param base
 * @returns {*}
 */
function customDivide(val, base) {

    if (val == undefined || val == '' || val == 0 || val == null) {
        return 0;
    }

    if (base == undefined || base == '' || base == 0 || base == null) {
        return "--";
    }

    try {
        val = parseFloat(val);
        base = parseFloat(base);
        return Math.round(val / base);
    } catch (e) {
        console.error("error:%s", e)
        return "--";
    }
}
/**
 * 自定义乘法
 * @param val
 * @param base
 * @returns {*}
 */
function customerMultiply(val, base, toFixed) {


    if (val == undefined || val == '' || val == 0 || val == null) {
        return 0;
    }

    if (base == undefined || base == '' || base == 0 || base == null) {
        return "--";
    }

    try {
        val = parseFloat(val);
        base = parseFloat(base);
        return (val * base).toFixed(parseInt(toFixed));
    } catch (e) {
        console.error("error:%s", e)
        return "--";
    }
}


/**
 * 自定义加法
 * @param val1
 * @param val2
 */
function customerAdd(val1, val2, op, toFixed) {

    if (val1 == undefined || val1 == '' || val1 == null) {
        return '--';
    }

    if (val2 == undefined || val2 == '' || val2 == null) {
        return '--';
    }

    if (op == undefined || op == '' || op == 0 || op == null) {
        return 0;
    }

    var ret = parseFloat(val1) + parseFloat(op) * parseFloat(val2);
    return ret.toFixed(toFixed);
}


/**
 * 涨停跌停价格计算
 * @param val 开盘价
 * @param op 1|-1
 */
function exPriceCal(val, op) {

    if (val == undefined || val == '' || val == 0 || val == null) {
        return 0;
    }

    if (op == undefined || op == '' || op == 0 || op == null) {
        return 0;
    }
    return customerMultiply(parseFloat(val), 1 + parseFloat(op));
}

/**
 *
 * 计算可买的 （手）
 * @param ava 可以金额
 * @param cp 当前价格
 */
function calcAvq(ava, cp, toFixed) {
    if (ava == undefined || ava == '' || ava == 0 || ava == null) {
        return 0;
    }
    if (cp == undefined || cp == '' || cp == 0 || cp == null) {
        return '-';
    }
    return (parseInt(ava) / parseInt(customerMultiply(cp, 100))).toFixed(toFixed)
}

/**
 * 
 * 格式化单位:
 * @param {*} val  100
 * @param {*} op 0:转化成手 ，1:转化成股
 * 
 * 1000,0: 10
 */
function formatHand(val, op) {

    if (val == undefined || val == '' || val == null) {
        return 0;
    }
    var x = parseInt(val);
    if (x % 100 != 0) {
         var d = Math.round(x/100)*100
         return d;
    }else{
        return x;
    }
}
