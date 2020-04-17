# frozen_string_literal: true

namespace :one_shot do
  task hard_delete: :environment do
    [
      Cast,
      Channel,
      ChannelGroup,
      Character,
      Collection,
      CollectionItem,
      Episode,
      EpisodeRecord,
      FaqCategory,
      FaqContent,
      Organization,
      Person,
      Program,
      Record,
      Series,
      SeriesWork,
      Slot,
      Staff,
      Trailer,
      User,
      VodTitle,
      Work,
      WorkRecord,
      WorkTag
    ].each do |model|
      puts "---------- model: #{model.name}"
      model.deleted.find_each do |record|
        puts "----- model: #{model.name}, record.id: #{record.id}"
        ActiveRecord::Base.connection.disable_referential_integrity do
          record.delete
        end
      end
    end
  end

  task update_name_on_series: :environment do
    Series.where.not(name_en: "").find_each do |series|
      puts "series.id: #{series.id}"
      series.update_column(:name_alter_en, series.name_en)
    end

    Series.where.not(name_ro: "").find_each do |series|
      puts "series.id: #{series.id}"
      series.update_column(:name_en, series.name_ro)
    end

    Series.where.not(name_alter_en: "").find_each do |series|
      puts "series.id: #{series.id}"
      if series.name_alter_en == series.name_en
        series.update_column(:name_alter_en, "")
      end
    end
  end
end
