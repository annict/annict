# typed: false
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
end
