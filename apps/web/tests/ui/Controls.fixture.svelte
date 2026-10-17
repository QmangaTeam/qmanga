<script lang="ts">
 import Button from '../../src/lib/shared/presentation/ui/Button.svelte';
 import ActionLink from '../../src/lib/shared/presentation/ui/ActionLink.svelte';
 import Icon from '../../src/lib/shared/presentation/ui/Icon.svelte';
 import IconButton from '../../src/lib/shared/presentation/ui/IconButton.svelte';
 import Field from '../../src/lib/shared/presentation/ui/Field.svelte';
 const unsafeText = '<script>unsafe()' + '<' + '/script>';
</script>

<Button data-case="default" name="choice" value="save">Сақтау</Button>
<Button data-case="submit" type="submit" form="sample-form">Жіберу</Button>
<Button data-case="busy" busy aria-label="Сақтау">Сақтау</Button>
<Button data-case="disabled" disabled>Келесі бет</Button>
<ActionLink href="/catalog" data-case="link">Каталог</ActionLink>
<Icon name="search"/>
<IconButton name="close" label="Жабу" data-case="icon-button"/>
<Field id="sample-email" label="Email" hint="Пошта мекенжайы" error="Мекенжайды тексер" required>
 {#snippet children(control)}<input type="email" {...control}/>{/snippet}
</Field>
<Field id="sample-select" label="Тіл" disabled>
 {#snippet children(control)}<select {...control}><option>Қазақша</option></select>{/snippet}
</Field>
<Field id="sample-note" label={unsafeText}>
 {#snippet children(control)}<textarea {...control}></textarea>{/snippet}
</Field>
