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
		var edgeLabels = $('#edgeLabels');
		edgeLabels.empty();
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
			for (var dstId in node.Edges) {
			  var edge = node.Edges[dstId];
				var spots = edge.Units.length;
				for (var i = 0; i < spots; i++) {
				  var spot = edge.Units[i];
					var text = document.createElementNS(SVG, 'text');
					var textPath = document.createElementNS(SVG, 'textPath');
					text.appendChild(textPath);
					textPath.setAttribute('xlink:href', '#' + edge.Src + '_' + edge.Dst + '_edge');
					textPath.setAttribute('startOffset', '' + (((i + 1) / (spots + 1)) * 100) + '%');
					var found = 0;
					for (var playerId in spot) {
					  found += 1;
					  var tspan = document.createElementNS(SVG, 'tspan');
						tspan.setAttribute('fill', players[playerId].color);
						tspan.setAttribute('font-weight', 'bold');
						tspan.textContent = '' + spot[playerId] + ' ';
						textPath.appendChild(tspan);
						console.log('found units');
					}
					if (found > 0) {
						edgeLabels[0].appendChild(text);
					}
				}
			}
		}
		var parentNode = that.$('svg').parent()[0];
		parentNode.innerHTML = parentNode.innerHTML;
	},

	prepareMap: function() {
	  var g = $('svg #graph0')[0];
		var edgeLabels = document.createElementNS(SVG, 'g');
		edgeLabels.setAttribute('id', 'edgeLabels');
		g.appendChild(edgeLabels);
	  $('g.edge path').each(function(i, path) {
		  path.setAttribute('id', $(path).closest('g.edge').attr('id') + '_edge');
		})
	},

  render: function() {
		var that = this;
    that.$el.html(that.template({
		  model: that.model,
		}));
		var playerNames = that.model.get('PlayerNames');
		if (playerNames != null) {
			that.prepareMap();
		  var colors = uniqueColors(playerNames.length);
			for (var i = 0; i < colors.length; i++) {
			  that.$('.players').append('<div style="color: ' + colors[i] + ';">' + playerNames[i] + ' </div>');
			}
		  that.renderTurn(that.currenTurn);
		}
		return that;
	},

});
