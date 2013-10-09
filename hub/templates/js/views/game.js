window.GameView = Backbone.View.extend({
	
	template: _.template($('#game_underscore').html()),

	initialize: function(options) {
	  this.model = new Game({}, {
		  url: '/games/' + options.id,
		});
		this.listenTo(this.model, 'change', this.render);
		this.model.fetch();
	},

  render: function() {
		var that = this;
    that.$el.html(that.template({
		  model: that.model,
		}));
		return that;
	},

});
