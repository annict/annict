class DB::MultipleEpisodesForm
  include ActiveModel::Model
  include Virtus.model
  include MultipleEpisodesFormatter

  attr_accessor :work

  attribute :body, String

  validates :body, presence: true

  def save
    return false unless valid?

    Episode.create_from_multiple_episodes(work, to_episode_hash)
  end
end
