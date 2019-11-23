# frozen_string_literal: true

namespace :tmp do
  task update_deleted_at: :environment do
    [
      Cast,
      Channel,
      Character,
      Collection,
      CollectionItem,
      Doorkeeper::Application,
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
      model.where(aasm_state: "hidden", deleted_at: nil).find_each do |record|
        puts "Update deleted_at on #{record.class.name} - id: #{record.id}"
        record.update_column(:deleted_at, record.updated_at)
      end
    end
  end
end
