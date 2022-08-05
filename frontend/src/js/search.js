
// MeiliSearch
const searchClient = new MeiliSearch({
	host: 'http://192.168.0.2:7700',
	apiKey: '',
})

const fp_index = searchClient.index('edea')

async function searchbox_handler() {
	var element = document.querySelector("input[name='searchbox']")
	const search = await fp_index.search(element.value)

	if (search.hits.length > 0) {
		var hits = document.querySelector("table[id='hits'] tbody")

		console.log(search.hits);

		var body = document.createElement('tbody');

		search.hits.forEach(e => {
			let row = body.insertRow(-1)
			let type = row.insertCell(0)
			let name = row.insertCell(1)
			let author = row.insertCell(2)
			let desc = row.insertCell(3)

			var link = document.createElement('a');
			link.innerHTML = e.Name

			if (e.Type == "module") {
				link.href = "/module/" + e.ID
			} else if (e.Type == "bench") {
				link.href = "/bench/" + e.ID
			}

			name.appendChild(link)
			type.innerHTML = e.Type
			author.innerHTML = e.Author
			desc.innerHTML = e.Description
		});

		// update the content with out search results
		hits.parentNode.replaceChild(body, hits)
	}
}

// if we have a query param, fill it out
var searchParams = new URLSearchParams(window.location.search);
if (searchParams.has('q')) {
	var element = document.querySelector("input[name='searchbox']");
	element.value = searchParams.get('q');
}

searchbox_handler();

function _create_button(button_label, aria_label, color) {
	let btn = document.createElement("button")
	btn.classList.add("btn", "btn-outline-" + color)
	btn.setAttribute("aria-label", aria_label)
	btn.innerHTML = button_label
	return btn
}

async function categories() {
	const categories = await fetch(`/api/search_fields`).then((response) => response.json())
	var filters_container = document.getElementById("filters-row")

	const cats = Object.keys(categories)

	if (cats.length > 0) {

		let i = 0;
		cats.forEach(cat => {
			let outer_div = document.createElement("div")
			outer_div.classList.add("col-3")  // change container width here when necessary

			let form_control_element_name = "filterf_" + cat

			let label = document.createElement('label')
			label.innerText = cat  // TODO add human readable category name here
			label.setAttribute("for", form_control_element_name)
			label.classList.add("form-label")

			outer_div.appendChild(label)

			let select = document.createElement('select')
			select.name = form_control_element_name
			select.setAttribute("multiple", "")
			select.setAttribute("aria-label", "filter options for " + cat)
			select.classList.add("form-select", "form-select-sm", "mb-1")

			// add an option for each value
			categories[cat].forEach(value => {
				let opt = document.createElement('option')
				opt.value = value
				opt.innerText = value
				select.appendChild(opt)
			})

			// add everything together
			outer_div.appendChild(select)

			let control_div = document.createElement("div")
			control_div.classList.add("input-group", "input-group-sm", "mb-3")

			control_div.appendChild(_create_button("&#x2264;", "select all values less than or equal to selected", "secondary"))
			let midbtn = _create_button("&#x21bb;", "clear this filter", "primary")
			midbtn.classList.add("form-control")
			control_div.appendChild(midbtn)
			control_div.appendChild(_create_button("&#x2265;", "select all values larger than or equal to selected", "secondary"))

			outer_div.appendChild(control_div)

			filters_container.appendChild(outer_div)
			i++
		})
	}

}

categories();

async function do_search() {
	var filter_row = document.getElementById("filter_row")

	var filter_ops = []

	for (var i = 0; i < filter_row.cells.length; i++) {
		let e = filter_row.cells[i].getElementsByTagName('select')[0]
		let collection = e.selectedOptions

		if (collection.length == 0) {
			continue
		}

		var op_values = []

		for (var j = 0; j < collection.length; j++) {
			op_values.add(collection[j].value)
		}

		filter_ops.add({ 'field': e.name, 'op': '=', 'values': op_values })
	}

	const results = await fetch(
		'/api/search_module',
		{ method: 'POST', body: JSON.stringify(filter_ops) }
	).then((response) => response.json())

	console.log(results)
	// TODO: display the results
}
