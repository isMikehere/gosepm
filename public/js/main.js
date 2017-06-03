/*
main.js
**/



$(document).ready(function(){
  
  //bind
  //商品选择
  $("#productType").bind("click",function(type){
      var url=$("#webpath").val()+"/product/"+$(this).val()
        $.Get(url,function(data){
            alert(data.data)
            // $("#price").html(data.Price)
        })
  });
  //提交订单 
  $("#buy").bind("click",function(){
      var url=$("#webpath").val()+"/user/follow"+$(this).val()
        $.Post(url,function(data){
        })
  })
//支付
$("#main").bind("click",function(){
     
     var url=$("#webpath").val()+"/pay"+$(this).val()
        $.Post(url,function(data){

       })
});


/**
 * 交易／获取股票信息
 */
$("#stockCode").bind("change",function (){
 var url=$("#webpath").val()+"/stock5/"+$(this).val()
        $.get(url,function(data){
          console.log(data)
       })
})

});