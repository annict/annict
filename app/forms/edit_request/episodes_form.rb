class EditRequest::EpisodesForm
  include Virtus.model
  include ActiveModel::Model

  attr_reader :edit_request_id, :user, :work

  attribute :episodes, String
  attribute :edit_request_title, String
  attribute :edit_request_body, String

  validates :episodes, presence: true

  def work=(work)
    @work ||= work
  end

  def user=(user)
    @user ||= user
  end

  def edit_request_id=(id)
    @edit_request_id ||= id
  end

  def attrs=(edit_request)
    self.edit_request_id = edit_request.id
    self.episodes = edit_request.draft_resource_params["episodes"]
    self.edit_request_title = edit_request.title
    self.edit_request_body = edit_request.body
  end

  def save
    return false unless valid?

    edit_request = if persisted?
      EditRequest.find(edit_request_id)
    else
      EditRequest.new
    end

    edit_request.attributes = {
      user: user,
      trackable: work,
      kind: :episodes,
      draft_resource_params: { episodes: episodes },
      title: edit_request_title.presence || "無題",
      body: edit_request_body
    }

    edit_request.save(validate: false)
    self.edit_request_id = edit_request.id
    true
  end

  def persisted?
    edit_request_id.present?
  end
end
