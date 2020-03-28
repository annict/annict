# frozen_string_literal: true

namespace :one_shot do
  task update_unpublished_at: :environment do
    [
      Series
    ].each do |model|
      model.deleted.find_each do |record|
        puts "model: #{model.name}, record.id: #{record.id}"
        record.update_column(:unpublished_at, record.deleted_at)
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
