# frozen_string_literal: true

namespace :tmp do
  task :move_org_from_people_to_orgs, [:person_id] => :environment do |_, args|
    person = Person.find(args[:person_id])
    org = Organization.where(name: person.name).first_or_create!

    person.staffs.find_each do |staff|
      puts "staff: #{staff.id}"
      org.staffs.where(work: staff.work, role: staff.role, role_other: staff.role_other).first_or_create! do |s|
        s.name = staff.name
        s.sort_number = staff.sort_number
      end
    end

    person.hide!
  end
end
