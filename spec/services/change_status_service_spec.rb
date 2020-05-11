# frozen_string_literal: true

describe ChangeStatusService, type: :service do
  context "when user does not add work to library entry" do
    let(:user) { create :registered_user }
    let(:work) { create :work }

    it "creates status" do
      expect(Status.count).to eq 0
      expect(Activity.count).to eq 0
      expect(LibraryEntry.count).to eq 0

      ChangeStatusService.new(user: user, work: work).call(status_kind: :watching)

      expect(Status.count).to eq 1
      expect(Activity.count).to eq 1
      expect(LibraryEntry.count).to eq 1

      status = user.statuses.first
      activity = user.activities.first
      library_entry = user.library_entries.first

      expect(status.kind).to eq "watching"
      expect(status.activity_id).to eq activity.id
      expect(status.work_id).to eq work.id

      expect(activity.action).to eq "create_status"
      expect(activity.recipient).to eq work
      expect(activity.trackable).to eq status
      expect(activity.single).to eq false
      expect(activity.repetitiveness).to eq false

      expect(library_entry.status_id).to eq status.id
      expect(library_entry.work_id).to eq work.id
    end
  end

  context "when user has added work to library entry" do
    let(:user) { create :registered_user }
    let(:episode) { create(:episode) }
    let(:work) { episode.work }
    let(:status) { create(:status, user: user, work: work, kind: :wanna_watch) }
    let!(:activity) { create(:activity, user: user, recipient: work, trackable: status, action: :create_status) }
    let!(:library_entry) { create(:library_entry, user: user, work: work, status: status) }

    it "creates status" do
      expect(Status.count).to eq 1
      expect(Activity.count).to eq 1
      expect(LibraryEntry.count).to eq 1
      expect(library_entry.status.kind).to eq "wanna_watch"

      ChangeStatusService.new(user: user, work: work).call(status_kind: :watching)

      expect(Status.count).to eq 2
      expect(Activity.count).to eq 2
      expect(LibraryEntry.count).to eq 1

      status = user.statuses.last
      activity_1 = user.activities.first
      activity_2 = user.activities.last
      library_entry = user.library_entries.first

      expect(status.kind).to eq "watching"
      expect(status.activity_id).to eq activity_1.id
      expect(status.work_id).to eq work.id

      expect(activity_2.action).to eq "create_status"
      expect(activity_2.recipient).to eq work
      expect(activity_2.trackable).to eq status
      expect(activity_2.single).to eq false
      expect(activity_2.repetitiveness).to eq true

      expect(library_entry.status_id).to eq status.id
      expect(library_entry.work_id).to eq work.id
    end
  end

  context "when user has added work to library entry and set no_select" do
    let(:user) { create :registered_user }
    let(:work) { create :work }
    let(:status) { create(:status, user: user, work: work, kind: :wanna_watch) }
    let!(:activity) { create(:activity, user: user, recipient: work, trackable: status, action: :create_status) }
    let!(:library_entry) { create(:library_entry, user: user, work: work, status: status) }

    it "resets status in library entry" do
      expect(Status.count).to eq 1
      expect(Activity.count).to eq 1
      expect(LibraryEntry.count).to eq 1
      expect(library_entry.status.kind).to eq "wanna_watch"

      ChangeStatusService.new(user: user, work: work).call(status_kind: "no_select")

      expect(Status.count).to eq 1
      expect(Activity.count).to eq 1
      expect(LibraryEntry.count).to eq 1

      library_entry = user.library_entries.first

      expect(library_entry.status).to be_nil
      expect(library_entry.work_id).to eq work.id
    end
  end
end
