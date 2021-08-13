# frozen_string_literal: true

namespace :data_care do
  task :merge_work, %i[base_work_id work_id] => :environment do |_, args|
    base_work_id, work_id = args.values_at(:base_work_id, :work_id)
    merge_work = Annict::DataCare::MergeWork.new(base_work_id, work_id)
    merge_work.run!
  end

  task :merge_episode, %i[base_episode_id episode_id] => :environment do |_, args|
    base_episode_id, episode_id = args.values_at(:base_episode_id, :episode_id)
    merge_episode = Annict::DataCare::MergeEpisode.new(base_episode_id, episode_id)
    merge_episode.run!
  end

  task :move_episode, %i[episode_id work_id] => :environment do |_, args|
    episode_id, work_id = args.values_at(:episode_id, :work_id)
    move_episode = Annict::DataCare::MoveEpisode.new(episode_id, work_id)
    move_episode.run!
  end

  task delete_abandoned_records: :environment do
    Activity.find_each do |a|
      if a.recipient.blank? || a.trackable.blank?
        puts "activity #{a.id} will be deleted"
        a.destroy
      end
    end
  end

  task :destroy_abandoned_activity_groups, [:user_id] => :environment do |_, args|
    User.find(args[:user_id]).activity_groups.find_each do |ag|
      if ag.activities.blank?
        puts "activity_groups.id: #{ag.id}"
        ag.destroy
      end
    end
  end

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

    person.destroy_in_batches
  end

  task :copy_casts, %i[base_work_id work_id] => :environment do |_, args|
    base_work = Work.find(args[:base_work_id])
    work = Work.find(args[:work_id])

    base_work.casts.order(:sort_number).each do |cast|
      work.casts.create(cast.attributes.except("id", "created_at", "updated_at"))
    end
  end

  task :copy_staffs, %i[base_work_id work_id] => :environment do |_, args|
    base_work = Work.find(args[:base_work_id])
    work = Work.find(args[:work_id])

    base_work.staffs.order(:sort_number).each do |staff|
      work.staffs.create(staff.attributes.except("id", "created_at", "updated_at"))
    end
  end

  task :set_sc_count_and_raw_number_to_works, %i[work_id] => :environment do |_, args|
    ActiveRecord::Base.transaction do
      work = Work.find(args[:work_id])
      work.episodes.order(:sort_number).each_with_index do |e, i|
        number = i + 1
        puts "#{e.number}: #{number}"
        e.update_columns(sc_count: number, raw_number: number)
      end
    end
  end

  task :reset_user_records_and_statuses, %i[username] => :environment do |_, args|
    user = User.only_kept.find_by(username: args[:username])

    puts "Deleting Records..."
    user.records.destroy_all

    puts "Deleting Multiple Records..."
    user.multiple_records.destroy_all

    puts "Deleting WorkRecords..."
    user.reviews.destroy_all

    puts "Deleting Statuses..."
    user.statuses.destroy_all
  end
end
