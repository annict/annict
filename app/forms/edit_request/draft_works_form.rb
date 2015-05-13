class EditRequest::DraftWorksForm
  include Virtus.model
  include ActiveModel::Model

  attribute :work_season_id, Integer
  attribute :work_sc_tid, Integer
  attribute :work_title, String
  attribute :work_media, String
  attribute :work_official_site_url, String
  attribute :work_wikipedia_url, String
  attribute :work_twitter_username, String
  attribute :work_twitter_hashtag, String
  attribute :work_released_at, String
  attribute :work_released_at_about, String
  attribute :work_fetch_syobocal, Axiom::Types::Boolean
  attribute :edit_request_title, String
  attribute :edit_request_body, String

  attribute :_edit_request_id, Integer, writer: :private
  attribute :_edit_request_user_id, Integer, writer: :private
  attribute :_edit_request_resource, Work, writer: :private

  attr_reader :work, :edit_request


  def edit_request_id=(id)
    self._edit_request_id = id
  end

  def edit_request_user_id=(user_id)
    self._edit_request_user_id = user_id
  end

  def edit_request_resource=(work)
    self._edit_request_resource = work
  end

  def work=(work)
    work_attributes.each do |attr|
      send("#{attr}=", work.send(attr.to_s.sub(/\Awork_/, "")))
    end
  end

  def edit_request=(edit_request)
    self.edit_request_id = edit_request.id
    self.edit_request_user_id = edit_request.user.id
    self.edit_request_resource = edit_request.resource

    work_attributes.each do |attr|
      send("#{attr}=", edit_request.draft_resource_params[attr.to_s.sub(/\Awork_/, "")])
    end

    self.edit_request_title = edit_request.title
    self.edit_request_body = edit_request.body

    self
  end

  def save
    return false unless valid?

    draft_resource_params = work_attributes.inject({}) do |hash, attr|
      hash.merge(attr.to_s.sub(/\Awork_/, "") => send(attr))
    end

    edit_request = EditRequest.where(id: _edit_request_id).first_or_initialize
    edit_request.attributes = {
      user_id: _edit_request_user_id,
      kind: :work,
      resource: _edit_request_resource,
      draft_resource_params: draft_resource_params,
      title: edit_request_title.presence || "無題",
      body: edit_request_body
    }

    edit_request.save!(validate: false)
    self.edit_request_id = edit_request.id

    true
  end

  def valid?
    build_resource

    work.valid?
    edit_request.valid?

    {}.merge(work.errors).merge(edit_request.errors).each do |key, errors|
      self.errors.add(key, *errors)
    end

    self.errors.blank?
  end

  def edit_request_persisted?
    _edit_request_id.present?
  end

  def create_path(controller)
    if _edit_request_resource.present?
      controller.db_work_edit_requests_path(_edit_request_resource)
    else
      controller.db_works_edit_requests_path
    end
  end

  private

  def build_resource
    @work = Work.new do |w|
      work_attributes.each do |attr|
        w.send("#{attr.to_s.sub(/\Awork_/, '')}=", send(attr))
      end
    end

    @edit_request = EditRequest.new do |er|
      er.title = edit_request_title
      er.body = edit_request_body
    end
  end

  def work_attributes
    attributes.keys.select { |attr| /\Awork_/ === attr }
  end
end
