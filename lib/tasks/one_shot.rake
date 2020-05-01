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

  task update_counter_cache: :environment do
    [
      [FavoriteCharacter, :favorite_characters_count, :favorite_characters],
      [FavoriteOrganization, :favorite_organizations_count, :favorite_organizations],
      [FavoritePerson, :favorite_people_count, :favorite_people]
    ].each do |(model, field, assoc)|
      User.only_kept.where(id: model.select(:user_id).distinct).find_each do |user|
        puts "#{model.name} > user: #{user.id}"
        user.update_column(field, user.send(assoc).size)
      end
    end
  end
end
