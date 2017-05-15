# frozen_string_literal: true

Types::WorkType = GraphQL::ObjectType.define do
  name "Work"
  description "An anime title"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  connection :episodes, Types::EpisodeType.connection_type do
    resolve ->(obj, _args, _ctx) {
      ForeignKeyLoader.for(Episode, :work_id).load([obj.id])
    }
  end

  field :annictId, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.id
    }
  end

  field :title, !types.String
  field :title_kana, types.String
  field :title_ro, types.String
  field :title_en, types.String

  field :media, !Types::MediaEnum do
    resolve ->(obj, _args, _ctx) {
      obj.media.upcase
    }
  end

  field :season_year, types.Int
  field :season_name, Types::SeasonNameEnum do
    resolve ->(obj, _args, _ctx) {
      obj.season_name&.upcase
    }
  end

  field :official_site_url, types.String
  field :official_site_url_en, types.String
  field :wikipedia_url, types.String
  field :wikipedia_url_en, types.String
  field :twitter_username, types.String
  field :twitter_hashtag, types.String

  field :image, Types::WorkImageType do
    resolve ->(obj, _args, _ctx) {
      ForeignKeyLoader.for(WorkImage, :work_id).load([obj.id])
    }
  end

  field :episodes_count, types.Int
  field :watchers_count, types.Int
end
