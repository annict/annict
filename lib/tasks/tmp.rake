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

  task hide_duplicate_resources: :environment do
    results = []

    Person.find_each do |p|
      person = Person.find_by(name: "#{p.name.strip} ")

      if person.present? && p.published? && person.published?
        puts "person.id: #{person.id}"

        person.casts.find_each do |c|
          c.update_column(:person_id, p.id)
        end

        person.staffs.find_each do |s|
          s.resource = p
          s.save!
        end

        person.hide!

        results << "p: #{p.id}, person: #{person.id}"
      end
    end

    Organization.find_each do |o|
      organization = Organization.find_by(name: "#{o.name.strip} ")

      if organization.present? && o.published? && organization.published?
        puts "organization.id: #{organization.id}"

        organization.staffs.find_each do |s|
          s.resource = o
          s.save!
        end

        organization.hide!

        results << "o: #{o.id}, organization: #{organization.id}"
      end
    end

    results.each { |r| puts r }
  end
end
