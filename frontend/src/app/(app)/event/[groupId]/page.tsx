interface EventGroupPageProps {
  params: Promise<{ groupId: string }>;
}

export default async function EventGroupPage({ params }: EventGroupPageProps) {
  const { groupId } = await params;

  return <div>Event {groupId}</div>;
}
