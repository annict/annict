# frozen_string_literal: true

describe ChangeStatusService, type: :service do
  describe "statuses" do
    context "when user does not add work to library entry" do
      let(:user) { create :registered_user }
      let(:work) { create :work }

      it "creates status" do
        expect(Status.count).to eq 0
        expect(Activity.count).to eq 0

        ChangeStatusService.new(user: user, work: work).call(status_kind: :watching)

        expect(Status.count).to eq 1
        expect(Activity.count).to eq 1

        status = user.statuses.where(work: work).first
        activity = user.activities.first

        expect(status.kind).to eq "watching"
        expect(status.activity_id).to eq activity.id
      end
    end

    context "when user has added work to library entry" do
      let(:user) { create :registered_user }
      let(:work) { create :work }
      let(:episode) { create(:episode, work: work) }
      let!(:library_entry) { create(:library_entry, user: user, work: work, next_episode: episode) }

      it "creates status" do
        expect(Status.count).to eq 1

        ChangeStatusService.new(user: user, work: work).call(status_kind: :watching)

        expect(Status.count).to eq 2

        status = user.statuses.where(work: work).last

        expect(status.kind).to eq "watching"
        expect(status.activity).not_to be_nil
      end
    end
  end

  describe "library_entries" do
    context "when user does not add work to library entry" do
      let(:user) { create :registered_user }
      let(:work) { create :work }

      it "creates library entry" do
        expect(LibraryEntry.count).to eq 0

        ChangeStatusService.new(user: user, work: work).call(status_kind: :watching)

        expect(LibraryEntry.count).to eq 1

        library_entry = user.library_entries.where(work: work).first

        expect(library_entry.status).not_to be_nil
      end
    end

    context "when user has added work to library entry" do
      let(:user) { create :registered_user }
      let(:work) { create :work }
      let(:status) { create(:status, user: user, work: work, kind: :wanna_watch) }
      let!(:library_entry) { create(:library_entry, user: user, work: work, status: status) }

      it "changes status in library entry" do
        expect(LibraryEntry.count).to eq 1
        expect(library_entry.status.kind).to eq "wanna_watch"

        ChangeStatusService.new(user: user, work: work).call(status_kind: :watching)

        expect(LibraryEntry.count).to eq 1

        library_entry = user.library_entries.where(work: work).first

        expect(library_entry.status.kind).to eq "watching"
      end
    end

    context "when user has added work to library entry and set no_select" do
      let(:user) { create :registered_user }
      let(:work) { create :work }
      let(:status) { create(:status, user: user, work: work, kind: :wanna_watch) }
      let!(:library_entry) { create(:library_entry, user: user, work: work, status: status) }

      it "resets status in library entry" do
        expect(LibraryEntry.count).to eq 1
        expect(library_entry.status.kind).to eq "wanna_watch"

        ChangeStatusService.new(user: user, work: work).call(status_kind: "no_select")

        expect(LibraryEntry.count).to eq 1

        library_entry = user.library_entries.where(work: work).first

        expect(library_entry.status).to be_nil
      end
    end
  end

  describe "activities" do
    let(:user) { create :registered_user }
    let(:work) { create :work }

    it "creates activity" do
      expect(Activity.count).to eq 0

      ChangeStatusService.new(user: user, work: work).call(status_kind: :watching)

      expect(Activity.count).to eq 1

      activity = user.activities.first

      expect(activity.action).to eq "create_status"
      expect(activity.resources_count).to eq 1
      expect(activity.single).to eq false
      expect(activity.trackable_type).to eq "Status"
    end
  end
end
