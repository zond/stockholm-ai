window.GameView = Backbone.View.extend({
	
	template: _.template($('#game_underscore').html()),

	events: {
	  'click .turn-forward-all': 'lastTurn',
	  'click .turn-back-all': 'firstTurn',
	  'click .turn-forward': 'nextTurn',
	  'click .turn-back': 'prevTurn',
	},

	initialize: function(options) {
	  this.model = new Game({}, {
		  url: '/games/' + options.id,
		});
		this.listenTo(this.model, 'change', this.render);
		this.model.fetch();
		this.currenTurn = 0;
	},

	firstTurn: function(ev) {
		ev.preventDefault();
		this.renderTurn(0);
	},

	lastTurn: function(ev) {
		ev.preventDefault();
		this.renderTurn(this.model.get('Turns').length - 1);
	},

	prevTurn: function(ev) {
		ev.preventDefault();
		this.renderTurn(this.currentTurn - 1);
	},

	nextTurn: function(ev) {
		ev.preventDefault();
		this.renderTurn(this.currentTurn + 1);
	},

  renderTurn: function(ordinal) {
	  var that = this;
		that.currentTurn = ordinal;
		that.$('.current-turn').val('' + ordinal);
		var turns = that.model.get('Turns');
	  if (ordinal == 0) {
		  that.$('.turn-back').attr('disabled', 'disabled'); 
		  that.$('.turn-back-all').attr('disabled', 'disabled'); 
		} else {
		  that.$('.turn-back').removeAttr('disabled');
		  that.$('.turn-back-all').removeAttr('disabled');
		}
		if (ordinal < turns.length - 1) {
		  that.$('.turn-forward').removeAttr('disabled');
		  that.$('.turn-forward-all').removeAttr('disabled');
		} else {
		  that.$('.turn-forward').attr('disabled', 'disabled'); 
		  that.$('.turn-forward-all').attr('disabled', 'disabled'); 
		}
	  var turn = turns[ordinal];
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
		  that.renderTurn(that.currenTurn);
		}
		return that;
	},

});
