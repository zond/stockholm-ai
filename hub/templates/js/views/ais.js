window.AIsView = Backbone.View.extend({
	
	template: _.template($('#ais_underscore').html()),

  render: function() {
		var that = this;
    that.$el.html(that.template({}));
		return that;
	},

});
