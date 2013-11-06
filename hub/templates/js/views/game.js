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
		this.currenTurn = options.ordinal || 0;
	},

	unlessFinished: function(cb) {
	  var that = this;
	  if (that.model.get('State') == 'Finished') {
		  cb();
		} else {
			that.model.fetch({
				success: cb,
			});
		}
	},

	firstTurn: function(ev) {
		ev.preventDefault();
		var that = this;
		that.unlessFinished(function() {
			that.renderTurn(0);
		});
	},

	lastTurn: function(ev) {
		ev.preventDefault();
		var that = this;
		that.unlessFinished(function() {
			that.renderTurn(that.model.get('Turns').length - 1);
		});
	},

	prevTurn: function(ev) {
		ev.preventDefault();
		var that = this;
		that.unlessFinished(function() {
			that.renderTurn(that.currentTurn - 1);
		});
	},

	nextTurn: function(ev) {
		ev.preventDefault();
		var that = this;
		that.unlessFinished(function() {
			that.renderTurn(that.currentTurn + 1);
		});
	},

  renderTurn: function(ordinal) {
	  var that = this;
		that.currentTurn = ordinal;
		that.$('.current-turn').attr('value', '' + ordinal);
		var turns = that.model.get('Length');
	  if (ordinal == 0) {
		  that.$('.turn-back').attr('disabled', 'disabled'); 
		  that.$('.turn-back-all').attr('disabled', 'disabled'); 
		} else {
		  that.$('.turn-back').removeAttr('disabled');
		  that.$('.turn-back-all').removeAttr('disabled');
		}
		if (ordinal < turns - 1) {
		  that.$('.turn-forward').removeAttr('disabled');
		  that.$('.turn-forward-all').removeAttr('disabled');
		} else {
		  that.$('.turn-forward').attr('disabled', 'disabled'); 
		  that.$('.turn-forward-all').attr('disabled', 'disabled'); 
		}
		window.session.router.navigate("/games/" + that.model.get('Id') + '/turns/' + ordinal);
	  var turnModel = new Turn({ url: '/games/' + that.model.get('Id') + '/turns/' + ordinal });
		turnModel.fetch({
			success: function() {
			  that.prepareMap(turnModel);
			  var turn = turnModel.attributes;
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
					var owners = [];
					for (var playerId in node.Units) {
						if (node.Units[playerId] > 0) {
							owners.push(playerId);
							var tspan = document.createElementNS(SVG, 'tspan');
							tspan.setAttribute('fill', players[playerId].color);
							tspan.setAttribute('font-weight', 'bold');
							tspan.textContent = '' + node.Units[playerId] + ' ';
							label[0].appendChild(tspan);
						}
					}
					if (owners.length == 1) {
						$('#' + selEscape(nodeId) + ' ellipse').first()[0].setAttribute('fill', players[owners[0]].color);
						var tspan = $('#' + selEscape(nodeId) + ' tspan').first()[0];
						tspan.setAttribute('fill', 'black');
					} else {
						$('#' + selEscape(nodeId) + ' ellipse').first()[0].setAttribute('fill', 'white');
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
								if (spot[playerId] > 0) {
									found += 1;
									var tspan = document.createElementNS(SVG, 'tspan');
									tspan.setAttribute('fill', players[playerId].color);
									tspan.setAttribute('font-weight', 'bold');
									tspan.textContent = '' + spot[playerId] + ' ';
									textPath.appendChild(tspan);
								}
							}
							if (found > 0) {
								edgeLabels[0].appendChild(text);
							}
						}
					}
					$('#' + selEscape(nodeId) + ' title').text('No change');
				}
				for (var nodeId in state.Changes) {
					var changes = state.Changes[nodeId];
					var messages = [];
					_.each(changes, function(change) {
						messages.push(players[change.PlayerId].name + ': ' + change.Units + ' (' + change.Reason + ')');
					});
					$('#' + selEscape(nodeId) + ' title').text(messages.join('\n'));
				}
				var parentNode = that.$('svg').parent()[0];
				parentNode.innerHTML = parentNode.innerHTML;
			},
		});
	},

	prepareMap: function(turn) {
	  if (this.$('svg').length == 0) {
		  this.$el.append(turn.svg());
			if (this.$('svg #edgeLabels').length == 0) {
				var g = $('svg #graph0')[0];
				$(g).find('polygon')[0].setAttribute('fill', $('body').css('background-color'));
				$('svg #graph0 > title').text('Nodes provide tooltips about what happened between this and last turn.');
				var edgeLabels = document.createElementNS(SVG, 'g');
				edgeLabels.setAttribute('id', 'edgeLabels');
				g.appendChild(edgeLabels);
				$('g.edge path').each(function(i, path) {
					path.setAttribute('id', $(path).closest('g.edge').attr('id') + '_edge');
				})
			}
		}
	},

  render: function() {
		var that = this;
		var playerNames = that.model.get('PlayerNames');
		if (playerNames != null) {
		  if (that.$('.players div').length == 0) {
				that.$el.html(that.template({
					model: that.model,
				}));
				var colors = uniqueColors(playerNames.length);
				for (var i = 0; i < colors.length; i++) {
					that.$('.players').append('<div style="color: ' + colors[i] + ';">' + playerNames[i] + ' </div>');
				}
				that.renderTurn(that.currenTurn);
			}
		}
		return that;
	},

});
