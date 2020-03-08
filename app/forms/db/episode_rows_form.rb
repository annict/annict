# frozen_string_literal: true

module DB
  class EpisodeRowsForm
    include ActiveModel::Model
    include Virtus.model
    include ResourceRows

    attribute :rows, String

    attr_accessor :user, :work

    validates :rows, presence: true

    def valid?
      super && new_episodes.all?(&:valid?)
    end

    def save!
      new_episodes_with_user.each(&:save_and_create_activity!)
    end

    private

    def attrs_list
      sort_number = @work.episodes.count * 100
      @attrs_list ||= parsed_rows.map do |row_columns|
        sort_number += 100
        {
          work_id: @work.id,
          number: row_columns[0],
          title: row_columns[1],
          sort_number: sort_number
        }
      end
    end

    def new_episodes
      @new_episodes ||= attrs_list.map { |attrs| Episode.new(attrs) }
    end

    def new_episodes_with_user
      @new_episodes_with_user ||= new_episodes.map do |episode|
        episode.user = @user
        episode
      end
    end
  end
end
