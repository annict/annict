# frozen_string_literal: true

namespace :counter_cache do
  this = self

  using Module.new {
    refine this.singleton_class do
      def clear_readonly_attributes!(model)
        model.class_eval do
          def self.readonly_attributes
            []
          end
        end
      end
    end
  }

  task refresh_all: :environment do
    [
      :refresh_on_characters,
      :refresh_on_episode_records,
      :refresh_on_episodes,
      :refresh_on_forum_posts,
      :refresh_on_multiple_episode_records,
      :refresh_on_organizations,
      :refresh_on_people,
      :refresh_on_series,
      :refresh_on_statuses,
      :refresh_on_userland_categories,
      :refresh_on_users,
      :refresh_on_work_records,
      :refresh_on_work_tags,
      :refresh_on_works,
    ].each do |task_name|
      puts "============== #{task_name} =============="
      Rake::Task["counter_cache:#{task_name}"].invoke
    end
  end

  task refresh_on_characters: :environment do
    clear_readonly_attributes!(Character)

    Character.only_kept.find_each do |character|
      favorite_users_count = character.users.only_kept.count

      next if character.favorite_users_count == favorite_users_count

      puts [
        "Character: #{character.id}",
        "favorite_users_count: #{character.favorite_users_count} -> #{favorite_users_count}"
      ].join(", ")

      character.update_columns(
        favorite_users_count: favorite_users_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_episode_records: :environment do
    clear_readonly_attributes!(EpisodeRecord)

    EpisodeRecord.only_kept.find_each do |episode_record|
      comments_count = episode_record.comments.count
      likes_count = episode_record.likes.count

      next if episode_record.comments_count == comments_count &&
        episode_record.likes_count == likes_count

      puts [
        "EpisodeRecord: #{episode_record.id}",
        "comments_count: #{episode_record.comments_count} -> #{comments_count}",
        "likes_count: #{episode_record.likes_count} -> #{likes_count}"
      ].join(", ")

      episode_record.update_columns(
        comments_count: comments_count,
        likes_count: likes_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_episodes: :environment do
    clear_readonly_attributes!(Episode)

    Episode.only_kept.find_each do |episode|
      episode_record_bodies_count = episode.episode_records.only_kept.with_body.count
      episode_records_count = episode.episode_records.only_kept.count

      next if episode.episode_record_bodies_count == episode_record_bodies_count &&
        episode.episode_records_count == episode_records_count

      puts [
        "Episode: #{episode.id}",
        "episode_record_bodies_count: #{episode.episode_record_bodies_count} -> #{episode_record_bodies_count}",
        "episode_records_count: #{episode.episode_records_count} -> #{episode_records_count}"
      ].join(", ")

      episode.update_columns(
        episode_record_bodies_count: episode_record_bodies_count,
        episode_records_count: episode_records_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_forum_posts: :environment do
    clear_readonly_attributes!(ForumPost)

    ForumPost.find_each do |forum_post|
      forum_comments_count = forum_post.forum_comments.count

      next if forum_post.forum_comments_count == forum_comments_count

      puts [
        "ForumPost: #{forum_post.id}",
        "forum_comments_count: #{forum_post.forum_comments_count} -> #{forum_comments_count}"
      ].join(", ")

      forum_post.update_columns(
        forum_comments_count: forum_comments_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_multiple_episode_records: :environment do
    clear_readonly_attributes!(MultipleEpisodeRecord)

    MultipleEpisodeRecord.find_each do |multiple_episode_record|
      likes_count = multiple_episode_record.likes.count

      next if multiple_episode_record.likes_count == likes_count

      puts [
        "MultipleEpisodeRecord: #{multiple_episode_record.id}",
        "likes_count: #{multiple_episode_record.likes_count} -> #{likes_count}"
      ].join(", ")

      multiple_episode_record.update_columns(
        likes_count: likes_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_organizations: :environment do
    clear_readonly_attributes!(Organization)

    Organization.only_kept.find_each do |organization|
      favorite_users_count = organization.organization_favorites.count
      staffs_count = organization.staffs.only_kept.count

      next if organization.favorite_users_count == favorite_users_count &&
        organization.staffs_count == staffs_count

      puts [
        "Organization: #{organization.id}",
        "favorite_users_count: #{organization.favorite_users_count} -> #{favorite_users_count}",
        "staffs_count: #{organization.staffs_count} -> #{staffs_count}"
      ].join(", ")

      organization.update_columns(
        favorite_users_count: favorite_users_count,
        staffs_count: staffs_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_people: :environment do
    clear_readonly_attributes!(Person)

    Person.only_kept.find_each do |person|
      casts_count = person.casts.only_kept.count
      favorite_users_count = person.person_favorites.count
      staffs_count = person.staffs.only_kept.count

      next if person.casts_count == casts_count &&
        person.favorite_users_count == favorite_users_count &&
        person.staffs_count == staffs_count

      puts [
        "Person: #{person.id}",
        "casts_count: #{person.casts_count} -> #{casts_count}",
        "favorite_users_count: #{person.favorite_users_count} -> #{favorite_users_count}",
        "staffs_count: #{person.staffs_count} -> #{staffs_count}"
      ].join(", ")

      person.update_columns(
        casts_count: casts_count,
        favorite_users_count: favorite_users_count,
        staffs_count: staffs_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_series: :environment do
    clear_readonly_attributes!(Series)

    Series.only_kept.find_each do |series|
      series_works_count = series.series_works.only_kept.count

      next if series.series_works_count == series_works_count

      puts [
        "Series: #{series.id}",
        "series_works_count: #{series.series_works_count} -> #{series_works_count}"
      ].join(", ")

      series.update_columns(
        series_works_count: series_works_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_statuses: :environment do
    clear_readonly_attributes!(Status)

    Status.find_each do |status|
      likes_count = status.likes.count

      next if status.likes_count == likes_count

      puts [
        "Status: #{status.id}",
        "likes_count: #{status.likes_count} -> #{likes_count}"
      ].join(", ")

      status.update_columns(
        likes_count: likes_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_userland_categories: :environment do
    clear_readonly_attributes!(UserlandCategory)

    UserlandCategory.find_each do |userland_category|
      userland_projects_count = userland_category.userland_projects.count

      next if userland_category.userland_projects_count == userland_projects_count

      puts [
        "UserlandCategory: #{userland_category.id}",
        "userland_projects_count: #{userland_category.userland_projects_count} -> #{userland_projects_count}"
      ].join(", ")

      userland_category.update_columns(
        userland_projects_count: userland_projects_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_users: :environment do
    clear_readonly_attributes!(User)

    User.only_kept.find_each do |user|
      episode_records_count = user.episode_records.only_kept.count
      notifications_count = user.notifications.unread.count
      records_count = user.records.only_kept.count

      next if user.episode_records_count == episode_records_count &&
        user.notifications_count == notifications_count &&
        user.records_count == records_count

      puts [
        "User: #{user.id}",
        "episode_records_count: #{user.episode_records_count} -> #{episode_records_count}",
        "notifications_count: #{user.notifications_count} -> #{notifications_count}",
        "records_count: #{user.records_count} -> #{records_count}"
      ].join(", ")

      user.update_columns(
        episode_records_count: episode_records_count,
        notifications_count: notifications_count,
        records_count: records_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_work_records: :environment do
    clear_readonly_attributes!(WorkRecord)

    WorkRecord.only_kept.find_each do |work_record|
      likes_count = work_record.likes.count

      next if work_record.likes_count == likes_count

      puts [
        "WorkRecord: #{work_record.id}",
        "likes_count: #{work_record.likes_count} -> #{likes_count}"
      ].join(", ")

      work_record.update_columns(
        likes_count: likes_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_work_tags: :environment do
    clear_readonly_attributes!(WorkTag)

    WorkTag.only_kept.find_each do |work_tag|
      work_taggings_count = work_tag.work_taggings.count

      next if work_tag.work_taggings_count == work_taggings_count

      puts [
        "WorkTag: #{work_tag.id}",
        "work_taggings_count: #{work_tag.work_taggings_count} -> #{work_taggings_count}"
      ].join(", ")

      work_tag.update_columns(
        work_taggings_count: work_taggings_count,
        updated_at: Time.zone.now
      )
    end
  end

  task refresh_on_works: :environment do
    clear_readonly_attributes!(Work)

    Work.only_kept.find_each do |work|
      kinds = %w(wanna_watch watching watched).freeze

      episodes_count = work.episodes.only_kept.count
      records_count = work.records.only_kept.count
      watchers_count = work.library_entries.with_status(*kinds).count
      work_records_count = work.work_records.only_kept.count
      work_records_with_body_count = work.work_records.only_kept.with_body.count

      next if work.episodes_count == episodes_count &&
        work.records_count == records_count &&
        work.watchers_count == watchers_count &&
        work.work_records_count == work_records_count &&
        work.work_records_with_body_count == work_records_with_body_count

      puts [
        "Work: #{work.id}",
        "episodes_count: #{work.episodes_count} -> #{episodes_count}",
        "records_count: #{work.records_count} -> #{records_count}",
        "watchers_count: #{work.watchers_count} -> #{watchers_count}",
        "work_records_count: #{work.work_records_count} -> #{work_records_count}",
        "work_records_with_body_count: #{work.work_records_with_body_count} -> #{work_records_with_body_count}"
      ].join(", ")

      work.update_columns(
        episodes_count: episodes_count,
        records_count: records_count,
        watchers_count: watchers_count,
        work_records_count: work_records_count,
        work_records_with_body_count: work_records_with_body_count,
        updated_at: Time.zone.now
      )
    end
  end
end
