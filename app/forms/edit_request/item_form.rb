class EditRequest::ItemForm
  include Virtus.model
  include ActiveModel::Model

  attr_reader :edit_request_id, :item, :user, :work

  attribute :item_name, String
  attribute :item_url, String
  attribute :edit_request_image_file, String
  attribute :edit_request_title, String
  attribute :edit_request_body, String

  def work=(work)
    @work ||= work
  end

  def item=(item)
    @item ||= item
  end

  def user=(user)
    @user ||= user
  end

  def edit_request_id=(id)
    @edit_request_id ||= id
  end

  def new_attributes=(program)
    self.program_channel_id = program.channel_id
    self.program_episode_id = program.episode_id
    self.program_started_at = program.started_at.to_time.strftime("%Y/%m/%d %H:%M")
  end

  def edit_attributes=(edit_request)
    self.edit_request_id = edit_request.id
    self.item_name = edit_request.draft_resource_params["name"]
    self.item_url = edit_request.draft_resource_params["url"]
    self.edit_request_title = edit_request.title
    self.edit_request_body = edit_request.body
  end

  def create_path(controller)
    if item.present?
      controller.db_work_item_edit_requests_path(work, item)
    else
      controller.db_work_items_edit_requests_path(work)
    end
  end

  def update_path(controller)
    if program.present?
      controller.db_work_program_edit_request_path(work, program, edit_request_id)
    else
      controller.db_work_programs_edit_request_path(work, edit_request_id)
    end
  end

  def valid?
    edit_request_image_file = if edit_request_image_file.blank?
      record = EditRequest.find(edit_request_id)
      record.images.first.image
    end

    item = Item.new do |i|
      i.work = work
      i.name = item_name
      i.url = item_url
      i.tombo_image = edit_request_image_file
    end

    item.valid?

    item.errors.each do |key, errors|
      self.errors.add("item_#{key}", *errors)
    end

    self.errors.blank?
  end

  def save
    return false unless valid?

    draft_resource_params = item_attrs.inject({}) do |hash, attr|
      hash.merge(attr => send("item_#{attr}"))
    end

    edit_request = if persisted?
      record = EditRequest.find(edit_request_id)
      record.images.destroy_all
      record
    else
      EditRequest.new
    end

    ActiveRecord::Base.transaction do
      edit_request.images.build(image: edit_request_image_file)
      edit_request.attributes = {
        user: user,
        trackable: work,
        kind: :item,
        resource: item,
        draft_resource_params: draft_resource_params,
        title: edit_request_title.presence || "無題",
        body: edit_request_body,
      }

      edit_request.save(validate: false)
      draft_resource_params = edit_request.draft_resource_params.merge(
        edit_request_image_id: edit_request.images.first.id
      )
      edit_request.update_column(:draft_resource_params, draft_resource_params)
    end

    self.edit_request_id = edit_request.id

    true
  end

  def persisted?
    edit_request_id.present?
  end

  private

  def item_attrs
    attributes.keys.select { |attr| /\Aitem_/ === attr }.map do |attr|
      attr.to_s.sub(/\Aitem_/, "")
    end
  end
end
