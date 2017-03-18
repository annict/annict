# frozen_string_literal: true

namespace :annict_db do
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

  task :copy_casts, %i(base_work_id work_id) => :environment do |_, args|
    base_work = Work.find(args[:base_work_id])
    work = Work.find(args[:work_id])

    base_work.casts.order(:sort_number).each do |cast|
      work.casts.create(cast.attributes.except("id", "created_at", "updated_at"))
    end
  end

  task :copy_staffs, %i(base_work_id work_id) => :environment do |_, args|
    base_work = Work.find(args[:base_work_id])
    work = Work.find(args[:work_id])

    base_work.staffs.order(:sort_number).each do |staff|
      work.staffs.create(staff.attributes.except("id", "created_at", "updated_at"))
    end
  end
end
