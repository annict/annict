# frozen_string_literal: true

namespace :seed do
  task generate_csv: :environment do
    Rake::Task["seed:generate_channel_groups_csv"].invoke
    Rake::Task["seed:generate_channels_csv"].invoke
    Rake::Task["seed:generate_works_csv"].invoke
    Rake::Task["seed:generate_number_formats_csv"].invoke
    Rake::Task["seed:generate_prefectures_csv"].invoke
    Rake::Task["seed:generate_seasons_csv"].invoke
    Rake::Task["seed:generate_tips_csv"].invoke
  end

  task generate_channel_groups_csv: :environment do
    attrs = %w(id sc_chgid name sort_number)

    CSV.open("#{Dir.pwd}/db/data/csv/channel_groups.csv", "wb") do |csv|
      csv << attrs

      ChannelGroup.select(attrs).find_each do |record|
        puts "ChannelGroup: #{record.id}"
        csv << attrs.map { |attr| record.send(attr) }
      end
    end
  end

  task generate_channels_csv: :environment do
    attrs = %w(id channel_group_id sc_chid name published)

    CSV.open("#{Dir.pwd}/db/data/csv/channels.csv", "wb") do |csv|
      csv << attrs

      Channel.select(attrs).find_each do |record|
        puts "Channel: #{record.id}"
        csv << attrs.map { |attr| record.send(attr) }
      end
    end
  end

  task generate_works_csv: :environment do
    work_attrs = %w(
      id season_id sc_tid title media official_site_url wikipedia_url released_at
      twitter_username twitter_hashtag released_at_about aasm_state number_format_id
      title_kana
    )

    works = []
    %w(
      ANNICT_PREVIOUS_SEASON
      ANNICT_CURRENT_SEASON
      ANNICT_NEXT_SEASON
    ).each do |const|
      year, name = ENV.fetch(const).split("-")
      works += Work.
        joins(:season).
        where(seasons: { year: year, name: name }).
        order(watchers_count: :desc).
        select(work_attrs).
        limit(10).
        to_a
    end
    works += Work.
      # To avoid `ActiveRecord::InvalidForeignKey` error
      # when run `rake db:seed`
      where.not(id: [865, 4370, 4447, 4177, 1540, 4266]).
      order(watchers_count: :desc).
      select(work_attrs).
      limit(100).
      to_a

    CSV.open("#{Dir.pwd}/db/data/csv/works.csv", "wb") do |csv|
      csv << work_attrs

      works.each do |record|
        puts "Work: #{record.id}"
        csv << work_attrs.map do |attr|
          case attr
          when "title_kana" then "-"
          else
            record.send(attr)
          end
        end
      end
    end

    episode_attrs = %w(
      id work_id number sort_number sc_count title prev_episode_id aasm_state
      fetch_syobocal raw_number
    )

    CSV.open("#{Dir.pwd}/db/data/csv/episodes.csv", "wb") do |csv|
      csv << episode_attrs

      works.each do |work|
        work.episodes.order(:id).select(episode_attrs).each do |record|
          puts "Episode: #{record.id}"
          csv << episode_attrs.map { |attr| record.send(attr) }
        end
      end
    end
  end

  task generate_number_formats_csv: :environment do
    attrs = %w(id name data sort_number format)

    CSV.open("#{Dir.pwd}/db/data/csv/number_formats.csv", "wb") do |csv|
      csv << attrs

      NumberFormat.select(attrs).find_each do |record|
        puts "NumberFormat: #{record.id}"
        csv << attrs.map { |attr| record.send(attr) }
      end
    end
  end

  task generate_prefectures_csv: :environment do
    attrs = %w(id name)

    CSV.open("#{Dir.pwd}/db/data/csv/prefectures.csv", "wb") do |csv|
      csv << attrs

      Prefecture.select(attrs).find_each do |record|
        puts "Prefecture: #{record.id}"
        csv << attrs.map { |attr| record.send(attr) }
      end
    end
  end

  task generate_seasons_csv: :environment do
    attrs = %w(id name sort_number year)

    CSV.open("#{Dir.pwd}/db/data/csv/seasons.csv", "wb") do |csv|
      csv << attrs

      Season.select(attrs).find_each do |record|
        puts "Season: #{record.id}"
        csv << attrs.map { |attr| record.send(attr) }
      end
    end
  end

  task generate_tips_csv: :environment do
    attrs = %w(id target slug title icon_name)

    CSV.open("#{Dir.pwd}/db/data/csv/tips.csv", "wb") do |csv|
      csv << attrs

      Tip.select(attrs).find_each do |record|
        puts "Tip: #{record.id}"
        csv << attrs.map { |attr| record.send(attr) }
      end
    end
  end
end
