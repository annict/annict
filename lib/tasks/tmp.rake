namespace :tmp do
  task remove_fetch_syobocal: :environment do
    DbActivity.where(trackable_type: "Work").where.not(action: "multiple_episodes.create").find_each do |dba|
      puts "id: #{dba.id}"
      begin
        dba.parameters["new"] = dba.parameters["new"].except("fetch_syobocal")
        if dba.parameters["old"].present?
          dba.parameters["old"] = dba.parameters["old"].except("fetch_syobocal")
        end
        dba.save!
      rescue
        binding.pry
      end
    end
  end
end
