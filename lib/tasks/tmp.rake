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
      Staff.create(work: wo.work, name: wo.organization.name,
        role: role, role_other: wo.role_other, sort_number: 200, resource: wo.organization)
    end
  end

  task :move_org_from_people_to_orgs, [:person_id, :org_id] => :environment do |_, args|
    person = Person.find(args[:person_id])
    org = Organization.find(args[:org_id])

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
