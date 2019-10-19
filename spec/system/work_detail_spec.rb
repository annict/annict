# frozen_string_literal: true

describe "Work detail page" do
  context "viewer does not sign in" do
    context "viewer's locale is `ja`" do
      before do
        allow_any_instance_of(Api::Internal::GraphqlController).to receive(:current_locale).and_return(:ja)
        allow_any_instance_of(Localable).to receive(:domain_jp?).and_return(true)
      end

      let!(:work) { create(:work, :with_current_season) }

      context "when any resources have not been added" do
        before do
          visit "/works/#{work.id}"
        end

        it "displays work data" do
          expect(page).to have_content(work.local_title)
          expect(page).to have_content(work.local_synopsis)
        end
      end

      context "when trailers have been added" do
        let!(:trailer) { create(:trailer, work: work) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays trailer title" do
          expect(page.find(".p-works-show__trailers")).to have_content(trailer.local_title)
        end
      end

      context "when episodes have been added" do
        let!(:episode) { create(:episode, work: work) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays episode title" do
          expect(page.find(".p-works-show__episodes")).to have_content(episode.local_title)
        end
      end

      context "when characters have been added" do
        let!(:cast) { create(:cast, work: work) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays character name" do
          expect(page.find(".p-works-show__characters")).to have_content(cast.character.local_name)
        end
      end

      context "when staffs (people) have been added" do
        let!(:person) { create(:person) }
        let!(:staff) { create(:staff, work: work, resource: person) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays staff name" do
          expect(page.find(".p-works-show__staffs")).to have_content(staff.resource.local_name)
        end
      end

      context "when staffs (organizations) have been added" do
        let!(:organization) { create(:organization) }
        let!(:staff) { create(:staff, work: work, resource: organization) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays staff name" do
          expect(page.find(".p-works-show__staffs")).to have_content(staff.resource.local_name)
        end
      end

      context "when vods have been added" do
        let!(:channel) { create(:channel, vod: true) }
        let!(:program) { create(:program, work: work, channel: channel, vod_title_code: "xxx") }
        let!(:vod_title_url) { "https://example.com/#{program.vod_title_code}" }

        before do
          allow_any_instance_of(Program).to receive(:vod_title_url).and_return(vod_title_url)

          visit "/works/#{work.id}"
        end

        it "can access to VOD service" do
          expect(page.find(".p-works-show__vods")).to have_link(href: vod_title_url)
        end
      end

      context "when work records have been added" do
        let!(:work_record) { create(:work_record, work: work) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays work record body" do
          expect(page.find(".p-works-show__work-records")).to have_content(work_record.body)
        end
      end

      context "when series have been added" do
        let!(:work2) { create(:work, :with_current_season) }
        let!(:series) { create(:series) }
        let!(:series_work) { create(:series_work, series: series, work: work) }
        let!(:series_work2) { create(:series_work, series: series, work: work2) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays series" do
          expect(page.find(".p-works-show__series")).to have_content(series_work.series.local_name)
          expect(page.find(".p-works-show__series")).to have_content(series_work.local_summary)
          expect(page.find(".p-works-show__series")).to have_link(href: "/works/#{work.id}")

          expect(page.find(".p-works-show__series")).to have_content(series_work2.series.local_name)
          expect(page.find(".p-works-show__series")).to have_content(series_work2.local_summary)
          expect(page.find(".p-works-show__series")).to have_link(href: "/works/#{work2.id}")
        end
      end
    end
  end

  context "viewer signs in" do
    let!(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    context "viewer's locale is `ja`" do
      before do
        allow_any_instance_of(Api::Internal::GraphqlController).to receive(:current_locale).and_return(:ja)
        allow_any_instance_of(Localable).to receive(:domain_jp?).and_return(true)
      end

      let!(:work) { create(:work, :with_current_season) }

      context "when any resources have not been added" do
        before do
          visit "/works/#{work.id}"
        end

        it "displays work data" do
          expect(page).to have_content(work.local_title)
          expect(page).to have_content(work.local_synopsis)
        end
      end

      context "when trailers have been added" do
        let!(:trailer) { create(:trailer, work: work) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays trailer title" do
          expect(page.find(".p-works-show__trailers")).to have_content(trailer.local_title)
        end
      end

      context "when episodes have been added" do
        let!(:episode) { create(:episode, work: work) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays episode title" do
          expect(page.find(".p-works-show__episodes")).to have_content(episode.local_title)
        end
      end

      context "when characters have been added" do
        let!(:cast) { create(:cast, work: work) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays character name" do
          expect(page.find(".p-works-show__characters")).to have_content(cast.character.local_name)
        end
      end

      context "when staffs (people) have been added" do
        let!(:person) { create(:person) }
        let!(:staff) { create(:staff, work: work, resource: person) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays staff name" do
          expect(page.find(".p-works-show__staffs")).to have_content(staff.resource.local_name)
        end
      end

      context "when staffs (organizations) have been added" do
        let!(:organization) { create(:organization) }
        let!(:staff) { create(:staff, work: work, resource: organization) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays staff name" do
          expect(page.find(".p-works-show__staffs")).to have_content(staff.resource.local_name)
        end
      end

      context "when vods have been added" do
        let!(:channel) { create(:channel, vod: true) }
        let!(:program) { create(:program, work: work, channel: channel, vod_title_code: "xxx") }
        let!(:vod_title_url) { "https://example.com/#{program.vod_title_code}" }

        before do
          allow_any_instance_of(Program).to receive(:vod_title_url).and_return(vod_title_url)

          visit "/works/#{work.id}"
        end

        it "can access to VOD service" do
          expect(page.find(".p-works-show__vods")).to have_link(href: vod_title_url)
        end
      end

      context "when work records have been added" do
        let!(:work_record) { create(:work_record, work: work) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays work record body" do
          expect(page.find(".p-works-show__work-records")).to have_content(work_record.body)
        end
      end

      context "when series have been added" do
        let!(:work2) { create(:work, :with_current_season) }
        let!(:series) { create(:series) }
        let!(:series_work) { create(:series_work, series: series, work: work) }
        let!(:series_work2) { create(:series_work, series: series, work: work2) }

        before do
          visit "/works/#{work.id}"
        end

        it "displays series" do
          expect(page.find(".p-works-show__series")).to have_content(series_work.series.local_name)
          expect(page.find(".p-works-show__series")).to have_content(series_work.local_summary)
          expect(page.find(".p-works-show__series")).to have_link(href: "/works/#{work.id}")

          expect(page.find(".p-works-show__series")).to have_content(series_work2.series.local_name)
          expect(page.find(".p-works-show__series")).to have_content(series_work2.local_summary)
          expect(page.find(".p-works-show__series")).to have_link(href: "/works/#{work2.id}")
        end
      end
    end
  end
end
