class DB::MultipleEpisodesForm
  include ActiveModel::Model
  include Virtus.model
  include DbActivityMethods
  include MultipleEpisodesFormatter

  attr_accessor :work

  attribute :body, String

  validates :body, presence: true
end
