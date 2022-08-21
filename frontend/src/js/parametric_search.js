// parametric module search

function select_range_handler(event) {
	event.preventDefault()

	let btn = event.srcElement
	let select_element = btn.parentElement.parentElement.children[1]

	if (btn.attributes["aria-label"].value.includes("less than")) {
		for (var i = 0; i < select_element.options.length; i++) {
			if (!select_element.options[i].selected) {
				select_element.options[i].selected = true
			} else {
				break
			}
		}
	} else if (btn.attributes["aria-label"].value.includes("larger than")) {
		for (var i = select_element.options.length - 1; i >= 0; i--) {
			if (!select_element.options[i].selected) {
				select_element.options[i].selected = true
			} else {
				break
			}
		}
	} else if (btn.attributes["aria-label"].value.includes("clear")) {
		for (var i = 0; i < select_element.options.length; i++) {
			if (select_element.options[i].selected) {
				select_element.options[i].selected = false
			}
		}
		btn.disabled = true
		btn.blur()
	}
}

async function async_add_module_to_bench(event) {
	event.preventDefault()
	const target_link = event.srcElement.href
	const button_elem = event.srcElement
	button_elem.blur()
	button_elem.disabled = true

	const response_ok = await fetch(target_link).then((response) => {
    if (!response.ok) {
      return false;
    }
    return true; }).catch((error) => {
    return false;
  });

	const module_counter = document.getElementById("modules-on-bench-counter")
	if (response_ok) {
		if (typeof(module_counter) != "undefined") {
			module_counter.innerText = 1 + Number(module_counter.innerText)
			const added_message = document.createElement("span")
			added_message.classList.add("badge", "bg-secondary")
			added_message.innerText = "Added!"
			button_elem.parentElement.insertBefore(added_message, button_elem.nextSibling)
		}
	} else {
		const error_message = document.createElement("span")
		error_message.classList.add("badge", "bg-danger")
		error_message.innerText = "Error"
		button_elem.parentElement.insertBefore(error_message, button_elem.nextSibling)
	}
}

function _create_button(button_label, aria_label, color, disabled = false) {
	let btn = document.createElement("button")
	btn.classList.add("btn")
	btn.setAttribute("aria-label", aria_label)
	btn.innerHTML = button_label
	if (!disabled) {
		btn.addEventListener('click', enable_update_filters_btn)
		btn.addEventListener('click', select_range_handler)
		btn.classList.add("btn-outline-" + color)
	} else {
		btn.classList.add("btn-outline-light")
		btn.setAttribute("disabled", true)
	}
	return btn
}

const filterfield_prefix = "filterf_"

async function categories() {
	const filters_container = document.getElementById("filters-row")
	if (typeof(filters_container) == "undefined") {
		return;
	}
	const filters = await fetch(`/api/filters`).then((response) => response.json())
	var filter_dict = {}
	for (var i = 0; i < filters.length; i++) {
		filter_dict[filters[i].Key] = filters[i]
	}
	const categories = await fetch(`/api/search_fields`).then((response) => response.json())
	const cats = Object.keys(categories)

	if (cats.length > 0) {

		let i = 0;
		cats.forEach(cat => {
			let outer_div = document.createElement("div")
			outer_div.classList.add("filterbox", "col-3")  // change container width here when necessary

			let form_control_element_name = filterfield_prefix + cat

			let label = document.createElement('label')
			var cat_name = cat
			var cat_label_found = false
			if (typeof (filter_dict[cat]) != 'undefined') {
				cat_name = filter_dict[cat].Name
				cat_label_found = true
			}
			label.innerText = cat_name
			label.setAttribute("for", form_control_element_name)
			label.classList.add("form-label")

			outer_div.appendChild(label)

			let select = document.createElement('select')
			select.name = form_control_element_name
			select.setAttribute("multiple", "")
			select.setAttribute("aria-label", "filter options for " + cat)
			select.classList.add("form-select", "form-select-sm", "mb-1")
			select.addEventListener('change', enable_update_filters_btn)

			// add an option for each value
			var num_cats = 0
			var num_selected = 0
			categories[cat].forEach(value => {
				let opt = document.createElement('option')
				opt.value = value
				opt.innerText = value
				select.appendChild(opt)
				num_cats++
				if (opt.selected) {
					num_selected++
				}
			})

			// add everything together
			outer_div.appendChild(select)

			let control_div = document.createElement("div")
			control_div.classList.add("input-group", "input-group-sm", "mb-3")

			control_div.appendChild(_create_button("&nbsp;&#x2264;&nbsp;", "select all values less than or equal to selected", "secondary", disabled = (num_cats < 2)))
			let midbtn = _create_button("&nbsp;&#x21bb;&nbsp;", "clear this filter", "primary", disabled = false)
			if (num_selected == 0) {
				midbtn.disabled = true
			}
			midbtn.classList.add("form-control")
			control_div.appendChild(midbtn)
			control_div.appendChild(_create_button("&nbsp;&#x2265;&nbsp;", "select all values larger than or equal to selected", "secondary", disabled = (num_cats < 2)))

			outer_div.appendChild(control_div)

			filters_container.appendChild(outer_div)
			i++
		})
	}

}

function disable_update_filters_btn() {
	update_filters_already_enabled = false
	let filter_button = document.getElementById("filter_apply_btn")
	filter_button.disabled = true
	filter_button.classList.add("btn-outline-light")
	filter_button.classList.remove("btn-primary")
	filter_button.removeEventListener('click', do_search)
}

let update_filters_already_enabled = false

function enable_update_filters_btn(event) {
	// enable reset filter button
	if (event.srcElement.nodeName == "SELECT") {
		event.srcElement.parentElement.lastChild.children[1].disabled = false
	}

	if (update_filters_already_enabled) {
		return
	}
	update_filters_already_enabled = true
	let filter_button = document.getElementById("filter_apply_btn")
	filter_button.disabled = false
	filter_button.classList.remove("btn-outline-light")
	filter_button.classList.add("btn-primary")
	filter_button.addEventListener('click', do_search)
}

categories();
let search_results = []

async function do_search() {
	disable_update_filters_btn()
	var filter_row = document.getElementsByClassName("filterbox")

	var filter_ops = []

	for (var i = 0; i < filter_row.length; i++) {
		let e = filter_row[i].getElementsByTagName('select')[0]
		let collection = e.selectedOptions

		if (collection.length == 0) {
			continue
		}

		var op_values = []

		for (var j = 0; j < collection.length; j++) {
			op_values.push(collection[j].value)
		}

		filter_ops.push({ 'field': e.name.substring(filterfield_prefix.length), 'op': '=', 'values': op_values })
	}

	const results = await fetch(
		'/api/search_module',
		{ method: 'POST', body: JSON.stringify(filter_ops) }
	).then((response) => response.json())

	search_results = results  // put it into a global for easier debugging from dev console
	let results_container = document.getElementById("hits-row")

	for (var i = 0; i < results_container.children.length - 1; i++) {
		results_container.removeChild(results_container.lastChild)
	}

	for (var i = 0; i < results.length; i++) {
		let r = results[i]
		let result_container_div = document.createElement("div")
		result_container_div.classList.add("col-12", "search-result", "mb-3")

		let result_card = document.createElement("div")
		result_card.classList.add("card")
		result_container_div.appendChild(result_card)

		let result_header = document.createElement("div")
		result_header.classList.add("card-header")
		result_header.innerHTML = "&lt; add tags or categories here maybe? &gt;" // TODO
		result_card.appendChild(result_header)

		let result_body = document.createElement("div")
		result_body.classList.add("card-body")
		result_card.appendChild(result_body)

		let result_title = document.createElement("h5")
		result_title.classList.add("card-title")
		result_title.innerHTML = r.Name
		if (r.User.Handle != "") {
			result_title.innerHTML += " <small> by <i>" + r.User.Handle + "</i></small>"
		}
		result_body.appendChild(result_title)

		let result_text = document.createElement("p")
		result_text.classList.add("card-text")
		result_text.innerHTML = r.Description
		result_body.appendChild(result_text)

		let module_link = document.createElement("a")
		module_link.classList.add("btn", "btn-outline-primary", "btn-sm", "mx-2")
		module_link.innerHTML = "Go to Module"  // consider loading it async in a pane?
		module_link.href = "/module/" + r.ID
		result_body.appendChild(module_link)

		let author_profile = document.createElement("a")
		author_profile.classList.add("btn", "btn-outline-primary", "btn-sm", "mx-2")
		author_profile.innerHTML = "Go to Author's Modules"
		author_profile.href = "/module/user/" + r.UserID
		result_body.appendChild(author_profile)

		let add_to_bench_link = document.createElement("button")
		add_to_bench_link.classList.add("btn", "btn-outline-secondary", "btn-sm", "mx-2")
		add_to_bench_link.innerHTML = "Add to my Bench"
		add_to_bench_link.href = "/bench/add/" + r.ID
		add_to_bench_link.addEventListener('click', async_add_module_to_bench)
		result_body.appendChild(add_to_bench_link) // TODO: make this asynchronous

		let result_footer = document.createElement("ul")
		result_footer.classList.add(
			"list-inline", "card-text", "text-muted", "small", "mt-3")
		const footer_contents = [
			"BOM: " + r.Metadata.count_part + " parts (" + r.Metadata.count_unique + " unique)",
			"PCB area: &lt;tbd&gt; mm&#xb2;",
			'last updated <time datetime="' + r.UpdatedAt + '">' + prettyDate(r.UpdatedAt) + "</time>"
		]
		footer_contents.forEach(entry => {
			let list_item = document.createElement("li")
			list_item.classList.add("list-inline-item")
			list_item.innerHTML = entry
			result_footer.appendChild(list_item)
		})
		result_body.appendChild(result_footer)

		results_container.appendChild(result_container_div)
	}
}
