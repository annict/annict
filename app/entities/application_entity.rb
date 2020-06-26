# frozen_string_literal: true

class ApplicationEntity < Dry::Struct
  schema schema.strict

  module Types
    include Dry.Types(default: :strict)

    ActivityResourceKinds = Types::String.enum("episode_record", "status", "work_record")
    RecordResourceKinds = Types::String.enum("episode_record", "work_record")
    WorkMediaKinds = Types::String.enum("tv", "ova", "movie", "web", "other")
    SeasonKinds = Types::String.enum("winter", "spring", "summer", "autumn")
    StatusKinds = Types::String.enum("plan_to_watch", "watching", "completed", "on_hold", "dropped", "no_status")
    RecordRatingStateKinds = Types::String.enum("great", "good", "average", "bad")
  end

  class << self
    private

    def local_attributes(*names)
      names.each do |name|
        define_method "local_#{name}" do
          value = send(name.to_sym)

          return send("#{name}_en".to_sym).presence || value if I18n.locale == :en

          value
        end
      end
    end
  end
end
