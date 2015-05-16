class EditRequest::WorkForm
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

  attr_reader :edit_request, :edit_request_id, :user, :work

  def edit_request_id=(id)
    @edit_request_id ||= id
  end

  def user=(user)
    @user ||= user
  end

  def work=(work)
    @work ||= work
  end

  def attrs=(work)
    work_attrs.each do |attr|
      send("work_#{attr}=", work.send(attr))
    end
  end

  def edit_request=(edit_request)
    self.edit_request_id = edit_request.id
    self.user = edit_request.user.id
    self.work = edit_request.resource

    work_attrs.each do |attr|
      send("work_#{attr}=", edit_request.draft_resource_params[attr])
    end

    self.edit_request_title = edit_request.title
    self.edit_request_body = edit_request.body

    self
  end

  def save
    return false unless valid?

    draft_resource_params = work_attrs.inject({}) do |hash, attr|
      hash.merge(attr => send("work_#{attr}"))
    end

    edit_request = if persisted?
      EditRequest.find(edit_request_id)
    else
      EditRequest.new
    end

    edit_request.attributes = {
      user: user,
      kind: :work,
      resource: work,
      draft_resource_params: draft_resource_params,
      title: edit_request_title.presence || "無題",
      body: edit_request_body
    }

    edit_request.save!(validate: false)
    self.edit_request_id = edit_request.id

    true
  end

  def valid?
    work_errors, edit_request_errors = validate_resources

    {}.merge(work_errors).merge(edit_request_errors).each do |key, errors|
      self.errors.add(key, *errors)
    end

    self.errors.blank?
  end

  def persisted?
    edit_request_id.present?
  end

  def create_path(controller)
    if work.present?
      controller.db_work_edit_requests_path(work)
    else
      controller.db_works_edit_requests_path
    end
  end

  private

  def validate_resources
    work = Work.new do |w|
      work_attrs.each do |attr|
        w.send("#{attr}=", send("work_#{attr}"))
      end
    end

    edit_request = EditRequest.new do |er|
      er.title = edit_request_title
      er.body = edit_request_body
    end

    work.valid?
    edit_request.valid?

    [work.errors, edit_request.errors]
  end

  def work_attrs
    attributes.keys.select { |attr| /\Awork_/ === attr }.map do |attr|
      attr.to_s.sub(/\Awork_/, "")
    end
  end
end
