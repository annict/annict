module StaffCommon
  extend ActiveSupport::Concern

  DIFF_FIELDS = %i(person_id name role role_other sort_number).freeze
  PUBLISH_FIELDS = (DIFF_FIELDS + %i(work_id)).freeze

  included do
    extend Enumerize
    enumerize :role, in: %w(
      original_creator
      chief_director
      director
      series_composition
      script
      original_character_design
      character_design
      chief_animation_director
      animation_director
      art_director
      sound_director
      music
      studio
      other
    )

    belongs_to :person
    belongs_to :resource, polymorphic: true
    belongs_to :work, touch: true

    validates :person_id, presence: true
    validates :work_id, presence: true
    validates :name, presence: true
    validates :role, presence: true

    scope :major, -> { where.not(role: "other") }

    def to_diffable_hash
      data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
        hash[field] = case field
        when :role
          send(field).to_s if send(field).present?
        else
          send(field)
        end

        hash
      end

      data.delete_if { |_, v| v.blank? }
    end
  end
end
