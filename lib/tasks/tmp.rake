namespace :tmp do
  task delete_main_attr: :environment do
    DbActivity.where(trackable_type: "Item").find_each do |a|
      new_params = a.parameters["new"]
      old_params = a.parameters["old"]

      parameters = {}
      if new_params.present?
        new_params.delete("main")
        parameters["new"] = new_params
      end
      if old_params.present?
        old_params.delete("main")
        parameters["old"] = old_params
      end

      a.update_column(:parameters, parameters)
      puts "done. #{a.id}"
    end
  end

  task delete_items_count: :environment do
    DbActivity.where(trackable_type: "Work").find_each do |w|
      new_params = w.parameters["new"]
      old_params = w.parameters["old"]

      parameters = {}
      if new_params.present?
        new_params.delete("items_count")
        parameters["new"] = new_params
      end
      if old_params.present?
        old_params.delete("items_count")
        parameters["old"] = old_params
      end

      w.update_column(:parameters, parameters)
      puts "done. #{w.id}"
    end
  end
end
