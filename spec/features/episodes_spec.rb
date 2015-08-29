describe 'エピソード詳細ページ' do
  let(:work)    { create(:work, :with_item) }
  let(:episode) { create(:episode, work: work) }

  before do
    visit "/works/#{work.id}/episodes/#{episode.id}"
  end

  it 'エピソード詳細ページが表示されること' do
    expect(page).to have_content(episode.title)
  end
end
