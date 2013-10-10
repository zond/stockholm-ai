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
		var players = {};
		var playerNames = that.model.get('PlayerNames');
		var playerIds = that.model.get('Players');
		var colors = uniqueColors(playerNames.length);
		for (var i = 0; i < playerIds.length; i++) {
		  players[playerIds[i]] = {
				name: playerNames[i],
				color: colors[i],
			};
		}
		for (var nodeId in state.Nodes) {
		  var node = state.Nodes[nodeId];
		  var label = $('#' + selEscape(nodeId) + ' text');
			label.empty();
      for (var playerId in node.Units) {
			  var tspan = document.createElementNS(SVG, 'tspan');
				tspan.setAttribute('fill', players[playerId].color);
				tspan.setAttribute('font-weight', 'bold');
				tspan.textContent = '' + node.Units[playerId] + ' ';
				label[0].appendChild(tspan);
			}
		}
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
			  that.$('.players').append('<div style="color: ' + colors[i] + ';">' + playerNames[i] + ' </div>');
			}
		  that.renderTurn(0);
		}
		return that;
	},

});
