var rem = {
	init:function(){
		this.resetRem();
		this.bindEvents();
	},
	
	resetRem:function(){
		var rect = window.document.documentElement.getBoundingClientRect();
        var width = rect.width > 1440 ? 1440 : rect.width;
        var rootEm = parseFloat(width/1440 * 20);
 
        document.getElementsByTagName('html')[0].style.fontSize=rootEm + "px";
	},
	
	bindEvents:function(){
		var self = this;
		window.onresize = function(){
		    window.setTimeout(self.resetRem, 300);
		}
	}
}

rem.init();