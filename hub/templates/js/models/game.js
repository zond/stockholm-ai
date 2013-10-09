window.Game = Backbone.Model.extend({

  idAttribute: 'Id',

	urlRoot: '/games',

	svg: function() {
	  var turns = this.get('Turns');
		if (turns != null) {
			var lastTurn = turns[turns.length - 1];
			var dot = [
				'digraph G {', 
					'graph [layout=neato,splines=true];',
			];
			for (var nodeId in lastTurn.State.Nodes) {
				var node = lastTurn.State.Nodes[nodeId];
				dot.push('node [shape=circle,label=' + node.Size + ']; "' + node.Id + '";');
				for (var edgeNodeId in node.Edges) {
					var edge = node.Edges[edgeNodeId];
					dot.push('"' + edge.Src + '" -> "' + edge.Dst + '"');
				}
			}
			console.log(dot.join('\n'));
			return Viz(dot.join('\n'), 'svg');
		} else {
		  return "";
		}
	},

});


