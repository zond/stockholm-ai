window.Turn = Backbone.Model.extend({

	idAttribute: 'Id',

	url: function() {
	  return this.url;
	},

	initialize: function(options) {
    this.url = options.url;
	},

	svg: function() {
		var state = this.get('State');
		var dot = [
			'digraph G {', 
				'graph [layout=neato,overlap=false,splines=true,sep="+5,5",size="20,10",ratio="fill"];',
		];
		var addedEdges = {};
		var edges = [];
		var nodes = [];
		for (var nodeId in state.Nodes) {
			var node = state.Nodes[nodeId];
			nodes.push(node);
			for (var edgeNodeId in node.Edges) {
				var edge = node.Edges[edgeNodeId];
				edges.push(edge);
			}
		}
		edges.sort(function(a, b) {
			if (a.edgeId > b.edgeId) {
				return 1;
			} else if (b.edgeId > a.edgeId) {
				return -1;
			} else {
				return 0;
			}
		});
		nodes.sort(function(a, b) {
			if (a.Id > b.Id) {
				return 1;
			} else if (b.Id > a.Id) {
				return -1;
			} else {
				return 0;
			}
		});
		_.each(nodes, function(node) {
			var size = node.Size / 50.0;
			dot.push('node [id="' + node.Id + '",shape=doublecircle,width="' + size + '",height="' + size + '",label=" "]; "' + node.Id + '";');
		});
		_.each(edges, function(edge) {
			dot.push('"' + edge.Src + '" -> "' + edge.Dst + '" [dir=none,id="' + edge.Src + '_' + edge.Dst + '",len="' + edge.Units.length + '"]');
		});
		dot.push('}');
		return Viz(dot.join('\n'), 'svg');
	},

});


