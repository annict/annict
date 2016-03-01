# frozen_string_literal: true

namespace :tmp do
  task update_resource: :environment do
    Staff.find_each do |s|
      puts "staff: #{s.id}"
      s.resource = s.person
      s.save!
    end
  end

  task add_wos_to_staffs: :environment do
    WorkOrganization.find_each do |wo|
      puts "wo: #{wo.id}"
      role = wo.role.value == "producer" ? :studio : :other
      Staff.create(person: Person.first, work: wo.work, name: wo.organization.name,
        role: role, role_other: wo.role_other, sort_number: 200, resource: wo.organization)
    end
  end
end
