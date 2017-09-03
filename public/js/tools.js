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
function customerMultiply(val, base) {


    if (val == undefined || val == '' || val == 0 || val == null) {
        return 0;
    }

    if (base == undefined || base == '' || base == 0 || base == null) {
        return "--";
    }

    try {
        val = parseFloat(val);
        base = parseFloat(base);
        return (val * base).toFixed(2);
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
function customerAdd(val1, val2, op) {

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
    console.log(ret)

    return ret.toFixed(2);
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







