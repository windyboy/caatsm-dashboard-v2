<script lang="ts">
  import { MessageDataTable } from "../MessageDataTable";
  import type { Telegram } from "$lib/types";

  interface MessagesSectionProps {
    messages: Telegram[];
    loading: boolean;
  }

  let { messages, loading }: MessagesSectionProps = $props();
</script>

<section class="messages-section">
  <MessageDataTable {messages} pageSize={10} {loading} />
</section>

<style>
  .messages-section {
    margin-bottom: 2rem;
  }
</style>
