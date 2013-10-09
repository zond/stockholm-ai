window.GameView = Backbone.View.extend({
	
	template: _.template($('#game_underscore').html()),

	initialize: function(options) {
	  this.model = new Game({}, {
		  url: '/games/' + options.id,
		});
		this.listenTo(this.model, 'change', this.render);
		this.model.fetch();
	},

  renderTurn: function(ordinal) {
	  var that = this;
	  var turn = that.model.get('Turns')[ordinal];
		var state = turn.State;
	},

  render: function() {
		var that = this;
    that.$el.html(that.template({
		  model: that.model,
		}));
		var playerNames = that.model.get('PlayerNames');
		if (playerNames != null) {
		  var colors = uniqueColors(playerNames.length);
			for (var i = 0; i < colors.length; i++) {
			  that.$('.players').append('<span class="player-name" style="color: ' + colors[i] + ';">' + playerNames[i] + ' </span>');
			}
		}
		if (that.model.get('Turns') != null) {
		  that.renderTurn(0);
		}
		return that;
	},

});
