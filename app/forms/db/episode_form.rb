# frozen_string_literal: true

module DB
  class EpisodeForm
    include ActiveModel::Model
    include Virtus.model
    include DbActivityMethods

    attribute :id, Integer
    attribute :raw_number, String
    attribute :number, String
    attribute :title, String
    attribute :title_ro, String
    attribute :title_en, String
    attribute :prev_episode_id, Integer
    attribute :sort_number, Integer
    attribute :sc_count, Integer
    attribute :fetch_syobocal, Boolean

    attr_accessor :root_resource
    attr_accessor :trackable_resource
    attr_accessor :new_record

    def valid?
      episode = Episode.new(to_h)

      return true if episode.valid?

      @errors = episode.errors
      false
    end

    def to_attributes
      to_h
    end

    def save!(*args)
      resource = Episode.where(id: to_attributes[:id]).first_or_initialize
      resource.attributes = to_attributes.except(:id)
      self.new_record = resource.new_record?
      resource.save!(args)
      self.trackable_resource = resource
    end
  end
end
