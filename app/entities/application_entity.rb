# frozen_string_literal: true

class ApplicationEntity < Dry::Struct
  module Types
    include Dry.Types(default: :strict)

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
          return send("#{name}_en".to_sym) if I18n.locale == :en

          send(name.to_sym)
        end
      end
    end
  end

  def decorate
    ActiveDecorator::Decorator.instance.decorate(self)
  end
end
