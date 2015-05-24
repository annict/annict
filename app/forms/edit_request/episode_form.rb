class EditRequest::EpisodeForm
  include Virtus.model
  include ActiveModel::Model

  attr_reader :edit_request_id, :episode, :user, :work

  attribute :episode_number, String
  attribute :episode_sort_number, Integer
  attribute :episode_title, String
  attribute :episode_next_episode_id, Integer
  attribute :edit_request_title, String
  attribute :edit_request_body, String

  def work=(work)
    @work ||= work
  end

  def episode=(episode)
    @episode ||= episode
  end

  def user=(user)
    @user ||= user
  end

  def edit_request_id=(id)
    @edit_request_id ||= id
  end

  def attrs=(resource)
    case resource.class.name
    when "Episode"
      self.episode_number = resource.number
      self.episode_sort_number = resource.sort_number
      self.episode_title = resource.title
      self.episode_next_episode_id = resource.next_episode_id
    when "EditRequest"
      self.edit_request_id = resource.id
      self.episode_number = resource.draft_resource_params["number"]
      self.episode_sort_number = resource.draft_resource_params["sort_number"]
      self.episode_title = resource.draft_resource_params["title"]
      self.episode_next_episode_id = resource.draft_resource_params["next_episode_id"]
      self.edit_request_title = resource.title
      self.edit_request_body = resource.body
    end
  end

  def valid?
    episode = Episode.new do |e|
      e.number = episode_number
      e.sort_number = episode_sort_number
      e.title = episode_title
      e.next_episode_id = episode_next_episode_id
    end

    episode.valid?

    {}.merge(episode.errors).each do |key, errors|
      self.errors.add(key, *errors)
    end

    self.errors.blank?
  end

  def save
    return false unless valid?

    draft_resource_params = episode_attrs.inject({}) do |hash, attr|
      hash.merge(attr => send("episode_#{attr}"))
    end

    edit_request = if persisted?
                     EditRequest.find(edit_request_id)
                   else
                     EditRequest.new
                   end

    edit_request.attributes = {
      user: user,
      trackable: work,
      kind: :episode,
      resource: episode,
      draft_resource_params: draft_resource_params,
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

  private

  def episode_attrs
    attributes.keys.select { |attr| /\Aepisode_/ === attr }.map do |attr|
      attr.to_s.sub(/\Aepisode_/, "")
    end
  end
end
